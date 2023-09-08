package model

type Rule struct {
	Name       string      `yaml:"name"`
	TriggerSet *TriggerSet `yaml:"triggerset"`
	Action     *Action     `yaml:"action"`
}
