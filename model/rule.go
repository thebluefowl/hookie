package model

type Rule struct {
	TriggerSet *TriggerSet `yaml:"triggers"`
	Action     *Action     `yaml:"action"`
}
