package segments

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/stretchr/testify/assert"
)

func TestMercurialEnabledToolNotFound(t *testing.T) {
	env := new(mock.Environment)
	env.On("InWSLSharedDrive").Return(false)
	env.On("HasCommand", "hg").Return(false)
	env.On("GOOS").Return("")
	env.On("IsWsl").Return(false)

	hg := &Mercurial{}
	hg.Init(properties.Map{}, env)

	assert.False(t, hg.Enabled())
}

func TestMercurialEnabledInWorkingDirectory(t *testing.T) {
	fileInfo := &runtime.FileInfo{
		Path:         "/dir/hello",
		ParentFolder: "/dir",
		IsDir:        true,
	}
	env := new(mock.Environment)
	env.On("InWSLSharedDrive").Return(false)
	env.On("HasCommand", "hg").Return(true)
	env.On("GOOS").Return("")
	env.On("IsWsl").Return(false)
	env.On("HasParentFilePath", ".hg", false).Return(fileInfo, nil)
	env.On("PathSeparator").Return("/")
	env.On("Home").Return(poshHome)
	env.On("Getenv", poshGitEnv).Return("")

	hg := &Mercurial{}
	hg.Init(properties.Map{}, env)

	assert.True(t, hg.Enabled())
	assert.Equal(t, fileInfo.Path, hg.mainSCMDir)
	assert.Equal(t, fileInfo.Path, hg.repoRootDir)
}

func TestMercurialGetIdInfo(t *testing.T) {
	cases := []struct {
		ExpectedWorking           *MercurialStatus
		Case                      string
		LogOutput                 string
		StatusOutput              string
		ExpectedBranch            string
		ExpectedChangeSetID       string
		ExpectedShortID           string
		ExpectedLocalCommitNumber string
		ExpectedBookmarks         []string
		ExpectedTags              []string
		ExpectedIsTip             bool
	}{
		{
			Case:         "nochanges_tip",
			LogOutput:    "123|b6cb23dcb79fe5c2215f1ae8f1a85326a7fed500|branchname|tip|",
			StatusOutput: "",
			ExpectedWorking: &MercurialStatus{ScmStatus{
				Modified:   0,
				Added:      0,
				Deleted:    0,
				Moved:      0,
				Untracked:  0,
				Conflicted: 0,
			}},
			ExpectedBranch:            "branchname",
			ExpectedChangeSetID:       "b6cb23dcb79fe5c2215f1ae8f1a85326a7fed500",
			ExpectedShortID:           "b6cb23dcb79f",
			ExpectedLocalCommitNumber: "123",
			ExpectedIsTip:             true,
			ExpectedBookmarks:         []string{},
			ExpectedTags:              []string{},
		},
		{
			Case:         "nochanges",
			LogOutput:    "123|b6cb23dcb79fe5c2215f1ae8f1a85326a7fed500|branchname||",
			StatusOutput: "",
			ExpectedWorking: &MercurialStatus{ScmStatus{
				Modified:   0,
				Added:      0,
				Deleted:    0,
				Moved:      0,
				Untracked:  0,
				Conflicted: 0,
			}},
			ExpectedBranch:            "branchname",
			ExpectedChangeSetID:       "b6cb23dcb79fe5c2215f1ae8f1a85326a7fed500",
			ExpectedShortID:           "b6cb23dcb79f",
			ExpectedLocalCommitNumber: "123",
			ExpectedIsTip:             false,
			ExpectedBookmarks:         []string{},
			ExpectedTags:              []string{},
		},
		{
			Case:      "changed",
			LogOutput: "3|11a953bf0288663b530dd6d65f3c8e0d5f7fddb5|default|tip mytag mytag2|bm1 bm2",
			StatusOutput: `
M Modified.File
? Untracked.File
R Removed.File
! AlsoRemoved.File
A Added.File
`,
			ExpectedWorking: &MercurialStatus{ScmStatus{
				Modified:   1,
				Added:      1,
				Deleted:    2,
				Moved:      0,
				Untracked:  1,
				Conflicted: 0,
			}},
			ExpectedBranch:            "default",
			ExpectedChangeSetID:       "11a953bf0288663b530dd6d65f3c8e0d5f7fddb5",
			ExpectedShortID:           "11a953bf0288",
			ExpectedLocalCommitNumber: "3",
			ExpectedIsTip:             true,
			ExpectedBookmarks:         []string{"bm1", "bm2"},
			ExpectedTags:              []string{"mytag", "mytag2"},
		},
	}

	for _, tc := range cases {
		fileInfo := &runtime.FileInfo{
			Path:         "/dir/hello",
			ParentFolder: "/dir",
			IsDir:        true,
		}

		props := properties.Map{
			FetchStatus: true,
		}

		env := new(mock.Environment)
		env.On("InWSLSharedDrive").Return(false)
		env.On("HasCommand", "hg").Return(true)
		env.On("GOOS").Return("")
		env.On("IsWsl").Return(false)
		env.On("HasParentFilePath", ".hg", false).Return(fileInfo, nil)
		env.On("PathSeparator").Return("/")
		env.On("Home").Return(poshHome)
		env.On("Getenv", poshGitEnv).Return("")
		env.MockHgCommand(fileInfo.Path, tc.LogOutput, "log", "-r", ".", "--template", hgLogTemplate)
		env.MockHgCommand(fileInfo.Path, tc.StatusOutput, "status")

		hg := &Mercurial{}
		hg.Init(props, env)

		if tc.ExpectedWorking != nil {
			tc.ExpectedWorking.Formats = map[string]string{}
		}

		assert.True(t, hg.Enabled())
		assert.Equal(t, fileInfo.Path, hg.mainSCMDir)
		assert.Equal(t, fileInfo.Path, hg.repoRootDir)
		assert.Equal(t, tc.ExpectedWorking, hg.Working, tc.Case)
		assert.Equal(t, tc.ExpectedBranch, hg.Branch, tc.Case)
		assert.Equal(t, tc.ExpectedChangeSetID, hg.ChangeSetID, tc.Case)
		assert.Equal(t, tc.ExpectedShortID, hg.ChangeSetIDShort, tc.Case)
		assert.Equal(t, tc.ExpectedLocalCommitNumber, hg.LocalCommitNumber, tc.Case)
		assert.Equal(t, tc.ExpectedIsTip, hg.IsTip, tc.Case)
		assert.Equal(t, tc.ExpectedBookmarks, hg.Bookmarks, tc.Case)
		assert.Equal(t, tc.ExpectedTags, hg.Tags, tc.Case)
	}
}
