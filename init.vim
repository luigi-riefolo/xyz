call plug#begin('~/.local/share/nvim/plugged')

"" Using a tagged release; wildcard allowed (requires git 1.9.2 or above)
Plug 'fatih/vim-go', { 'do': ':GoInstallBinaries' }

Plug 'modille/groovy.vim'

Plug 'jparise/vim-graphql'

"Plug 'artur-shaik/vim-javacomplete2'

Plug 'flazz/vim-colorschemes'

Plug 'scrooloose/nerdtree'

Plug 'stamblerre/gocode', { 'rtp': 'vim', 'do': '~/.vim/plugged/gocode/vim/symlink.sh' }

if has('nvim')
  Plug 'Shougo/deoplete.nvim', { 'do': ':UpdateRemotePlugins' }
  "Plug 'deoplete-plugins/deoplete-go', { 'do': 'make'}
else
  Plug 'Shougo/deoplete.nvim'
  Plug 'roxma/nvim-yarp'
  Plug 'roxma/vim-hug-neovim-rpc'
endif

Plug 'tpope/vim-fugitive'

Plug 'nvie/vim-flake8'

Plug 'maksimr/vim-jsbeautify'

Plug 'mhinz/vim-grepper'

Plug 'romainl/vim-qf'

Plug 'neomake/neomake'

" (Optinal) for Tag Sidebar
Plug 'majutsushi/tagbar'

Plug 'hashivim/vim-terraform'
Plug 'vim-syntastic/syntastic'
"Plug 'juliosueiras/vim-terraform-completion'

Plug 'Xuyuanp/nerdtree-git-plugin'

Plug 'airblade/vim-gitgutter'

Plug '907th/vim-auto-save'

Plug 'rust-lang/rust.vim'

"Plug 'dense-analysis/ale'

" Initialize plugin system
call plug#end()


let g:deoplete#enable_at_startup = 1
let g:deoplete#omni_patterns = {}
call deoplete#initialize()
"let g:deoplete#omni_patterns.terraform = '[^ *\t"{=$]\w*'
"call deoplete#custom#option('omni_patterns', {
"\ 'complete_method': 'omnifunc',
"\ 'terraform': '[^ *\t"{=$]\w*',
"\})


set directory=$HOME/.vim/swapfiles/

" add vertical lines on columns
set colorcolumn=80,120

" Show line numbers - could be toggled on/off on-fly by pressing F6
set number

" Ignore case when searching
set ignorecase

" When searching try to be smart about cases
set smartcase

" No annoying sound on errors
set noerrorbells
set novisualbell
set t_vb=
set tm=500

colorscheme basic-dark

" highlight trailing space
highlight ExtraWhitespace ctermbg=red guibg=red
match ExtraWhitespace /\s\+$/
autocmd BufWinEnter * match ExtraWhitespace /\s\+$/
autocmd InsertEnter * match ExtraWhitespace /\s\+\%#\@<!$/
autocmd InsertLeave * match ExtraWhitespace /\s\+$/
autocmd BufWinLeave * call clearmatches()

" Return to last edit position when opening files (You want this!)
autocmd BufReadPost *
     \ if line("'\"") > 0 && line("'\"") <= line("$") |
     \   exe "normal! g`\"" |
     \ endif

" Remove trailing spaces
autocmd BufWritePre * :%s/\s\+$//e

command! JSONpretty %!python -m json.tool
command! XMLpretty %!xmllint --format -
set mouse=a
let g:NERDTreeMouseMode=3

runtime nerdtree-plugin.vim
"set foldmethod=syntax

set tabstop=4
set shiftwidth=4
set softtabstop=4
set expandtab
set smarttab
set textwidth=79
set autoindent
set smartindent

set mmp=5000

autocmd StdinReadPre * let s:std_in=1
autocmd VimEnter * if argc() == 0 && !exists("s:std_in") | NERDTree | exe 'NERDTreeProjectLoad base' | endif
autocmd StdinReadPre * let s:std_in=1
autocmd VimEnter * if argc() == 1 && isdirectory(argv()[0]) && !exists("s:std_in") | exe 'NERDTree' argv()[0] | wincmd p | ene | exe 'NERDTreeProjectLoad base' | endif
"autocmd VimEnter * NERDTreeProjectLoad base
autocmd VimLeave * NERDTreeProjectSave base

let NERDTreeMinimalUI=1
let NERDTreeWinSize = 35
let NERDTreeShowHidden=1
let NERDTreeDirArrows=1
" Make sure that when NT root is changed, Vim's pwd is also updated
let NERDTreeChDirMode = 2
"let NERDTreeShowLineNumbers = 1
let NERDTreeAutoCenter = 1
let mapleader=","
" Locate file in hierarchy quickly
map <leader>j :NERDTreeFind<cr>
" Toogle on/off
nmap <leader>o :NERDTreeToggle<cr>


"vmap <C-x> :!pbcopy<CR>
"
"
"map <C-a> ggVG
"vmap <C-c> :w !xclip
"vmap <C-c> "+y
"map <C-c> "+y
vmap <C-c> y:call system("xclip -i -selection clipboard", getreg("\""))<CR>:call system("xclip -i", getreg("\""))<CR>
nmap <C-v> :call setreg("\"",system("xclip -o -selection clipboard"))<CR>p
set clipboard=unnamed

let g:go_metalinter_autosave = 1
" TODO: wait until they resolve the issue with errors not showing
let g:go_metalinter_command='golangci-lint'
"let g:go_metalinter_autosave_enabled = ['deadcode', 'errcheck', 'gosimple', 'govet', 'staticcheck', 'golint', 'typecheck', 'unused', 'varcheck']
"let g:go_metalinter_enabeld = ['deadcode', 'errcheck', 'gosimple', 'govet', 'staticcheck', 'typecheck', 'unused', 'varcheck']
"let g:go_metalinter_autosave_enabled = ['govet']
"let g:go_metalinter_command='gometalinter'
let g:go_fmt_command = "goimports"
"let g:go_list_type = "quickfix"
let g:ale_linters = {'go': ['golangci-lint']}


autocmd BufWritePre *.json :%!python -m json.tool


function! Smart_TabComplete()
  let line = getline('.')                         " current line

  let substr = strpart(line, -1, col('.')+1)      " from the start of the current
                                                  " line to one character right
                                                  " of the cursor
  let substr = matchstr(substr, "[^ \t]*$")       " word till cursor
  if (strlen(substr)==0)                          " nothing to match on empty string
    return "\<tab>"
  endif
  let has_period = match(substr, '\.') != -1      " position of period, if any
  let has_slash = match(substr, '\/') != -1       " position of slash, if any
  if (!has_period && !has_slash)
    return "\<C-X>\<C-P>"                         " existing text matching
  elseif ( has_slash )
    return "\<C-X>\<C-F>"                         " file matching
  else
    return "\<C-X>\<C-O>"                         " plugin matching
  endif
endfunction
inoremap <tab> <c-r>=Smart_TabComplete()<CR>

let NERDTreeWinSize=48

autocmd FileType yaml setlocal ts=2 sts=2 sw=2 expandtab
autocmd FileType yml setlocal ts=2 sts=2 sw=2 expandtab

" no split
set textwidth=0 wrapmargin=0
