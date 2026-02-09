package commands

import (
	"fmt"

	"github.com/FujishigeTemma/git-review/internal/output"
)

// SkillMarkdown holds the embedded SKILL.md text.
type SkillMarkdown string

type SkillCmd struct{}

func (c *SkillCmd) Run(content SkillMarkdown, out *output.Output) error {
	fmt.Fprint(out.Stdout, string(content))
	return nil
}
