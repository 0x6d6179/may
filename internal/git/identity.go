package git

import "fmt"

// SetLocalIdentity sets user.name and user.email in the local git config of dir.
func SetLocalIdentity(r *Runner, dir, name, email string) error {
	if _, err := r.RunInDir(dir, "config", "--local", "user.name", name); err != nil {
		return fmt.Errorf("set local identity name: %w", err)
	}
	if _, err := r.RunInDir(dir, "config", "--local", "user.email", email); err != nil {
		return fmt.Errorf("set local identity email: %w", err)
	}
	return nil
}

// GetLocalIdentity reads user.name and user.email from the git config in dir.
func GetLocalIdentity(r *Runner, dir string) (name, email string, err error) {
	name, err = r.RunInDir(dir, "config", "user.name")
	if err != nil {
		return "", "", fmt.Errorf("get local identity name: %w", err)
	}
	email, err = r.RunInDir(dir, "config", "user.email")
	if err != nil {
		return "", "", fmt.Errorf("get local identity email: %w", err)
	}
	return name, email, nil
}
