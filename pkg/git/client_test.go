package git

import (
	"regexp"
	"testing"
)

// TODO: make this test not depend on external resource
func TestGetRefs(t *testing.T) {
	// Run

	url := "https://github.com/makerdao/dss-deploy-scripts"
	refs, err := GetRefs(url)
	if err != nil {
		t.Error(err)
	}

	t.Logf("%v", refs)

	// Assertions
	if len(refs) == 0 {
		t.Errorf("No refs returned from ls-remote")
	}

	matchRev := regexp.MustCompile(`[a-f0-9]{40}`)
	matchRef := regexp.MustCompile(`refs/[-./_\w{}^~]+|HEAD`)
	for i, commit := range refs {
		t.Logf("ref:%s rev:%s", commit.Ref, commit.Rev)
		if commit.URL != url {
			t.Errorf("Commit nr %d's URL doesn't match requested: %s", i, url)
		}
		if !matchRev.MatchString(commit.Rev) {
			t.Errorf("Commit nr %d's Rev doesn't look like a hash: %s", i, commit.Rev)
		}
		if !matchRef.MatchString(commit.Ref) {
			t.Errorf("Commit nr %d's Ref doesn't look like a ref: %s", i, commit.Ref)
		}
	}
}

// TODO: make this test not depend on external resource
//func testGetRepoPath(t *testing.T) {
//	// Run
//	repoPath, err := GetRepoPath(Commit{
//		URL: "https://github.com/makerdao/dss-deploy-scripts",
//		Ref: "refs/tags/staxx-deploy",
//		Rev: "4b3682baa91ad34898c9fc8474b27d5721ff7150",
//	})
//	if err != nil {
//		t.Error(err)
//	}
//
//	t.Logf("Repo path: %s", repoPath)
//
//	// Assertions
//	_, err = os.Stat(repoPath)
//	if err != nil {
//		t.Errorf("Repo path doesn't exist")
//	}
//
//	_, err = os.Stat(repoPath)
//	if err != nil {
//		t.Errorf("Repo path doesn't exist")
//	}
//}
