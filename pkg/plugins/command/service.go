package command

import (
	"encoding/json"
	"sd/pkg/types"
)

// CommandPlugin represents the command plugin.
type CommandPlugin struct{}

// ExecuteAction implements actions.Plugin.
func (c *CommandPlugin) ExecuteAction(actionType types.ActionType, config json.RawMessage) error {
	panic("unimplemented")
}

// GetActionTypes implements actions.Plugin.
func (c *CommandPlugin) GetActionTypes() []types.ActionType {
	panic("unimplemented")
}

// ValidateConfig implements actions.Plugin.
func (c *CommandPlugin) ValidateConfig(actionType types.ActionType, config json.RawMessage) error {
	panic("unimplemented")
}

// Name returns the name of the plugin.
func (c *CommandPlugin) Name() string {
	return "command"
}

// Subscribe sets up the NATS subscription for this plugin.
func (c *CommandPlugin) Init() {
	OpenSubscriber()
}
