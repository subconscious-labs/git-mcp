// Package mcp defines the tool contract exposed by the server.
// Every struct in this file maps one-to-one to a tool of the Python
// reference mcp-server-git implementation, including required fields,
// optional fields and their defaults.
//
// GitFetchParams, GitPullParams and GitPushParams are extensions over the
// reference; they are only registered in --mode=remote-write.
package mcp

// GitStatusParams mirrors the reference GitStatus model.
type GitStatusParams struct {
	RepoPath string `json:"repo_path" jsonschema:"Path to the Git repository.,required"`
}

// GitDiffUnstagedParams mirrors GitDiffUnstaged.
type GitDiffUnstagedParams struct {
	RepoPath     string `json:"repo_path" jsonschema:"Path to the Git repository.,required"`
	ContextLines *int   `json:"context_lines,omitempty" jsonschema:"Number of context lines to show in diff output. Defaults to 3."`
}

// GitDiffStagedParams mirrors GitDiffStaged.
type GitDiffStagedParams struct {
	RepoPath     string `json:"repo_path" jsonschema:"Path to the Git repository.,required"`
	ContextLines *int   `json:"context_lines,omitempty" jsonschema:"Number of context lines to show in diff output. Defaults to 3."`
}

// GitDiffParams mirrors GitDiff.
type GitDiffParams struct {
	RepoPath     string `json:"repo_path" jsonschema:"Path to the Git repository.,required"`
	Target       string `json:"target" jsonschema:"Branch, tag, or commit to diff against.,required"`
	ContextLines *int   `json:"context_lines,omitempty" jsonschema:"Number of context lines to show in diff output. Defaults to 3."`
}

// GitCommitParams mirrors GitCommit.
type GitCommitParams struct {
	RepoPath string `json:"repo_path" jsonschema:"Path to the Git repository.,required"`
	Message  string `json:"message" jsonschema:"Commit message.,required"`
}

// GitAddParams mirrors GitAdd.
type GitAddParams struct {
	RepoPath string   `json:"repo_path" jsonschema:"Path to the Git repository.,required"`
	Files    []string `json:"files" jsonschema:"Files to stage.,required,minItems=1"`
}

// GitResetParams mirrors GitReset.
type GitResetParams struct {
	RepoPath string `json:"repo_path" jsonschema:"Path to the Git repository.,required"`
}

// GitLogParams mirrors GitLog.
type GitLogParams struct {
	RepoPath       string  `json:"repo_path" jsonschema:"Path to the Git repository.,required"`
	MaxCount       *int    `json:"max_count,omitempty" jsonschema:"Maximum number of commits to return. Defaults to 10."`
	StartTimestamp *string `json:"start_timestamp,omitempty" jsonschema:"Start timestamp for filtering commits."`
	EndTimestamp   *string `json:"end_timestamp,omitempty" jsonschema:"End timestamp for filtering commits."`
}

// GitCreateBranchParams mirrors GitCreateBranch.
type GitCreateBranchParams struct {
	RepoPath   string  `json:"repo_path" jsonschema:"Path to the Git repository.,required"`
	BranchName string  `json:"branch_name" jsonschema:"Name of the new branch.,required"`
	BaseBranch *string `json:"base_branch,omitempty" jsonschema:"Optional base branch. Defaults to the current HEAD."`
}

// GitCheckoutParams mirrors GitCheckout.
type GitCheckoutParams struct {
	RepoPath   string `json:"repo_path" jsonschema:"Path to the Git repository.,required"`
	BranchName string `json:"branch_name" jsonschema:"Branch or revision to check out.,required"`
}

// GitShowParams mirrors GitShow.
type GitShowParams struct {
	RepoPath string `json:"repo_path" jsonschema:"Path to the Git repository.,required"`
	Revision string `json:"revision" jsonschema:"Commit, tag, object, or <commit>:<path> to show.,required"`
}

// GitBranchParams mirrors GitBranch.
type GitBranchParams struct {
	RepoPath    string  `json:"repo_path" jsonschema:"Path to the Git repository.,required"`
	BranchType  string  `json:"branch_type" jsonschema:"One of: local, remote, all.,required,enum=local,enum=remote,enum=all"`
	Contains    *string `json:"contains,omitempty" jsonschema:"Only list branches containing this commit SHA."`
	NotContains *string `json:"not_contains,omitempty" jsonschema:"Only list branches not containing this commit SHA."`
}

// GitFetchParams is the extension tool git_fetch. Not in the Python
// reference; only registered in --mode=remote-write.
type GitFetchParams struct {
	RepoPath string `json:"repo_path" jsonschema:"Path to the Git repository.,required"`
	Remote   string `json:"remote,omitempty" jsonschema:"Remote name. Defaults to origin."`
}

// GitPullParams is the extension tool git_pull. Not in the Python
// reference; only registered in --mode=remote-write.
type GitPullParams struct {
	RepoPath string `json:"repo_path" jsonschema:"Path to the Git repository.,required"`
	Remote   string `json:"remote,omitempty" jsonschema:"Remote name. Defaults to origin."`
	Branch   string `json:"branch,omitempty" jsonschema:"Branch to pull. Defaults to the remote-tracking branch of the current HEAD."`
}

// GitPushParams is the extension tool git_push. Not in the Python
// reference; only registered in --mode=remote-write.
type GitPushParams struct {
	RepoPath string `json:"repo_path" jsonschema:"Path to the Git repository.,required"`
	Remote   string `json:"remote" jsonschema:"Remote name to push to. Defaults to origin."`
	Branch   string `json:"branch" jsonschema:"Branch or refspec to push. Defaults to the current branch."`
}

// DefaultContextLines matches DEFAULT_CONTEXT_LINES in the reference.
const DefaultContextLines = 3

// DefaultMaxCount matches the default max_count of git_log.
const DefaultMaxCount = 10
