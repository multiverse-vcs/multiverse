package repo

import (
	"github.com/multiverse-vcs/go-git-ipfs/internal/core"
)

// RepoLogsPerPage is the amount of logs per page.
const RepoLogsPerPage = 30

const (
	RepoInfoTab = "info"
	RepoTreeTab = "tree"
	RepoRefsTab = "refs"
	RepoLogsTab = "logs"
)

type Repo core.Server
