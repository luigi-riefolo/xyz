package api

import (
	"context"
	"fmt"
	"log"
	"time"

	"cloud.google.com/go/firestore"
	"firebase.google.com/go/auth"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"github.com/pkg/errors"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/luigi-riefolo/xyz/pb"
)

// Device represents a device that can be used in a project.
type Device struct {
	ID          string
	Description string
}

// Project represents an XYZ project.
type Project struct {
	CreatorID string

	Contributors []string
	Devices      map[string]Device

	CreatedAt int64
	UpdatedAt int64
}

// CreateUser creates a user.
func (s *Service) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.User, error) {

	params := (&auth.UserToCreate{}).
		Email(req.Email).
		EmailVerified(true).
		Password(req.Password).
		DisplayName(fmt.Sprintf("%s %s", req.Firstname, req.Lastname)).
		Disabled(false)

	fbUser, err := s.authClient.CreateUser(ctx, params)
	if err != nil {
		if userExists := auth.IsEmailAlreadyExists(err); userExists {
			return nil, errors.Wrapf(err, "user '%s' already exists", req.Email)
		}
		return nil, errors.Wrapf(err, "could not create user '%s'", req.Email)
	}

	log.Printf("successfully created user '%s'\n", fbUser.UserInfo.UID)

	// TODO: add all metadata and info that can be returned to user
	user := &pb.User{
		Id: fbUser.UserInfo.UID,
	}

	return user, nil
}

// CreateProject creates a project.
func (s *Service) CreateProject(ctx context.Context, req *pb.CreateProjectRequest) (*pb.Project, error) {

	contributors := []string{}

	for _, email := range req.Contributors {
		if _, err := s.authClient.GetUserByEmail(ctx, email); err != nil {
			if auth.IsUserNotFound(err) {
				log.Printf("user '%s' does not exist, skipping\n", email)
			} else {
				return nil, errors.Wrap(err, "could not create project")
			}
			continue
		}
		contributors = append(contributors, email)
	}

	// TODO: add all project properties
	project := Project{
		CreatorID:    metautils.ExtractIncoming(ctx).Get(userEmailHeader),
		Contributors: contributors,
		CreatedAt:    time.Now().Unix(),
	}

	doc, _, err := s.projects.Add(ctx, project)
	if err != nil {
		return nil, errors.Wrap(err, "could not create project")
	}

	ret := &pb.Project{
		Id:           doc.ID,
		CreatorId:    project.CreatorID,
		Contributors: contributors,
		CreatedAt:    project.CreatedAt,
	}

	log.Printf("successfully created project '%s'\n", doc.ID)

	return ret, nil
}

// GetProjects returns the list of projects.
func (s *Service) GetProjects(ctx context.Context, req *empty.Empty) (*pb.ProjectsList, error) {

	iter := s.projects.Documents(ctx)

	ret := &pb.ProjectsList{
		List: []*pb.Project{},
	}

	defer iter.Stop()
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, errors.Wrap(err, "could not list projects")
		}

		project := &pb.Project{}

		if err := doc.DataTo(project); err != nil {
			return nil, errors.Wrap(err, "could not unmarshal projects")
		}

		ret.List = append(ret.List, project)
	}

	return ret, nil
}

// AddProjectContributors adds one or more contributors to a project.
func (s *Service) AddProjectContributors(ctx context.Context, req *pb.AddProjectContributorsRequest) (*empty.Empty, error) {

	doc := s.projects.Doc(req.ProjectId)

	projectData, err := doc.Get(ctx)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("id was not found: %v", err))
	}

	project := &pb.Project{}
	if err := projectData.DataTo(project); err != nil {
		return nil, errors.Wrap(err, "could not unmarshal projects")
	}

	if project.CreatorId != metautils.ExtractIncoming(ctx).Get(userEmailHeader) {
		return nil, errors.Wrap(err, "user is not project owner")
	}

	contributors := []string{}
	for _, email := range req.Contributors {
		if _, err := s.authClient.GetUserByEmail(ctx, email); err != nil {
			if auth.IsUserNotFound(err) {
				log.Printf("user '%s' does not exist, skipping\n", email)
			} else {
				return nil, errors.Wrap(err, "could not validate contributor")
			}
			continue
		}
		contributors = append(contributors, email)
	}

	project.Contributors = append(project.Contributors, contributors...)

	_, err = doc.Update(ctx,
		[]firestore.Update{
			{
				Path:  "Contributors",
				Value: project.Contributors,
			},
			{
				Path:  "UpdatedAt",
				Value: time.Now().Unix(),
			},
		})
	if err != nil {
		return nil, errors.Wrap(err, "could not update project's contributors")
	}

	log.Printf("successfully added contributors '%v' to project '%s'\n",
		project.Contributors,
		req.ProjectId)

	return &empty.Empty{}, nil
}

func userIsProjectContributor(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// AddProjectDevices adds one or more devices to a project.
func (s *Service) AddProjectDevices(ctx context.Context, req *pb.AddProjectDevicesRequest) (*empty.Empty, error) {

	doc := s.projects.Doc(req.ProjectId)

	projectData, err := doc.Get(ctx)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("id was not found: %v", err))
	}

	project := &pb.Project{}
	if err := projectData.DataTo(project); err != nil {
		return nil, errors.Wrap(err, "could not unmarshal projects")
	}

	user := metautils.ExtractIncoming(ctx).Get(userEmailHeader)
	if project.CreatorId != user && userIsProjectContributor(project.Contributors, user) {
		return nil, errors.Wrap(err, "user is not authorised")
	}

	project.Devices = append(project.Devices, req.Devices...)

	_, err = doc.Update(ctx,
		[]firestore.Update{
			{
				Path:  "Devices",
				Value: project.Devices,
			},
			{
				Path:  "UpdatedAt",
				Value: time.Now().Unix(),
			},
		})
	if err != nil {
		return nil, errors.Wrap(err, "could not update project's devices")
	}

	log.Printf("successfully added devices '%v' to project '%s'\n",
		req.Devices,
		req.ProjectId)

	return &empty.Empty{}, nil
}

// GetDevices returns the list of devices assigned to a project.
func (s *Service) GetDevices(ctx context.Context, req *pb.GetProjectDevicesRequest) (*pb.DevicesList, error) {

	doc := s.projects.Doc(req.ProjectId)

	projectData, err := doc.Get(ctx)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("id was not found: %v", err))
	}

	project := &pb.Project{}
	if err := projectData.DataTo(project); err != nil {
		return nil, errors.Wrap(err, "could not unmarshal projects")
	}

	user := metautils.ExtractIncoming(ctx).Get(userEmailHeader)
	if project.CreatorId != user && userIsProjectContributor(project.Contributors, user) {
		return nil, errors.Wrap(err, "user is not authorised")
	}

	return &pb.DevicesList{Devices: project.Devices}, nil
}
