package services

import "context"

type stubRepoManager struct{}

func (stubRepoManager) Kind() RepoKind                     { return RepoKindGit }
func (stubRepoManager) RootPath() string                   { return "" }
func (stubRepoManager) WorkspacePath(taskID string) string { return "" }
func (stubRepoManager) CreateWorkspace(ctx context.Context, taskID, name string) (string, error) {
	return "", nil
}
func (stubRepoManager) RemoveWorkspace(ctx context.Context, taskID string) error { return nil }
func (stubRepoManager) Commit(ctx context.Context, taskID, message string) error { return nil }
func (stubRepoManager) CommitMergeResolution(ctx context.Context, taskID, message string) error {
	return nil
}
func (stubRepoManager) AbortMerge(ctx context.Context, taskID string) error { return nil }
func (stubRepoManager) GetCurrentBranch(ctx context.Context, taskID string) (string, error) {
	return "", nil
}
func (stubRepoManager) PushBranch(ctx context.Context, taskID string) error        { return nil }
func (stubRepoManager) GetDiff(ctx context.Context, taskID string) (string, error) { return "", nil }
func (stubRepoManager) MergeToMain(ctx context.Context, taskID string) (string, error) {
	return "", nil
}
