package model

type RulesetAction struct {
	Ruleset *Ruleset `yaml:"ruleset"`
	Action  *Action  `yaml:"action"`
}


