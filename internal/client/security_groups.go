package client

import (
	"context"
	"fmt"
	"strconv"
)

type SecurityGroup struct {
	ID          int64               `json:"id"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Active      *bool               `json:"active"`
	Enabled     *bool               `json:"enabled"`
	Rules       []SecurityGroupRule `json:"rules"`
}

type SecurityGroupRule struct {
	ID                   int64  `json:"id"`
	Name                 string `json:"name"`
	Direction            string `json:"direction"`
	Policy               string `json:"policy"`
	Protocol             string `json:"protocol"`
	PortRange            string `json:"portRange"`
	DestinationPortRange string `json:"destinationPortRange"`
	SourceType           string `json:"sourceType"`
	Source               string `json:"source"`
	DestinationType      string `json:"destinationType"`
	Destination          string `json:"destination"`
	Ethertype            string `json:"ethertype"`
	Priority             *int64 `json:"priority"`
	Enabled              *bool  `json:"enabled"`
}

type SecurityGroupRuleInput struct {
	Name                 string
	Direction            string
	Policy               string
	Protocol             string
	PortRange            string
	DestinationPortRange string
	SourceType           string
	Source               string
	DestinationType      string
	Destination          string
	Ethertype            string
	Priority             *int64
	Enabled              *bool
}

func (c *Client) CreateSecurityGroup(ctx context.Context, name, description string) (*SecurityGroup, error) {
	payload := map[string]any{"securityGroup": map[string]any{"name": name}}
	if description != "" {
		payload["securityGroup"].(map[string]any)["description"] = description
	}
	var response struct {
		SecurityGroup SecurityGroup `json:"securityGroup"`
	}
	if err := c.post(ctx, "/security-groups", payload, &response); err != nil {
		return nil, err
	}
	return &response.SecurityGroup, nil
}

func (c *Client) GetSecurityGroup(ctx context.Context, id int64) (*SecurityGroup, error) {
	var response struct {
		SecurityGroup SecurityGroup `json:"securityGroup"`
	}
	if err := c.get(ctx, fmt.Sprintf("/security-groups/%d", id), nil, &response); err != nil {
		return nil, err
	}
	return &response.SecurityGroup, nil
}

func (c *Client) GetSecurityGroupByName(ctx context.Context, name string) (*SecurityGroup, error) {
	var response struct {
		SecurityGroups []SecurityGroup `json:"securityGroups"`
	}
	if err := c.get(ctx, "/security-groups", map[string]string{"name": name, "max": "1"}, &response); err != nil {
		return nil, err
	}
	if len(response.SecurityGroups) == 0 {
		return nil, &APIError{StatusCode: 404, Message: fmt.Sprintf("security group %q not found", name)}
	}
	return &response.SecurityGroups[0], nil
}

func (c *Client) UpdateSecurityGroup(ctx context.Context, id int64, name, description string) (*SecurityGroup, error) {
	payload := map[string]any{"securityGroup": map[string]any{"name": name, "description": description}}
	var response struct {
		SecurityGroup SecurityGroup `json:"securityGroup"`
	}
	if err := c.put(ctx, fmt.Sprintf("/security-groups/%d", id), payload, &response); err != nil {
		return nil, err
	}
	return &response.SecurityGroup, nil
}

func (c *Client) DeleteSecurityGroup(ctx context.Context, id int64) error {
	return c.delete(ctx, fmt.Sprintf("/security-groups/%d", id), nil)
}

func (c *Client) CreateSecurityGroupRule(ctx context.Context, securityGroupID int64, input SecurityGroupRuleInput) (*SecurityGroupRule, error) {
	var response struct {
		Rule SecurityGroupRule `json:"rule"`
	}
	if err := c.post(ctx, fmt.Sprintf("/security-groups/%d/rules", securityGroupID), mapRulePayload(input, false), &response); err != nil {
		return nil, err
	}
	return &response.Rule, nil
}

func (c *Client) UpdateSecurityGroupRule(ctx context.Context, securityGroupID, ruleID int64, input SecurityGroupRuleInput) (*SecurityGroupRule, error) {
	var response struct {
		Rule SecurityGroupRule `json:"rule"`
	}
	if err := c.put(ctx, fmt.Sprintf("/security-groups/%d/rules/%d", securityGroupID, ruleID), mapRulePayload(input, true), &response); err != nil {
		return nil, err
	}
	return &response.Rule, nil
}

func (c *Client) DeleteSecurityGroupRule(ctx context.Context, securityGroupID, ruleID int64) error {
	return c.delete(ctx, fmt.Sprintf("/security-groups/%d/rules/%d", securityGroupID, ruleID), nil)
}

func (c *Client) GetSecurityGroupRule(ctx context.Context, securityGroupID, ruleID int64) (*SecurityGroupRule, error) {
	sg, err := c.GetSecurityGroup(ctx, securityGroupID)
	if err != nil {
		return nil, err
	}
	for _, rule := range sg.Rules {
		if rule.ID == ruleID {
			return &rule, nil
		}
	}
	return nil, &APIError{StatusCode: 404, Message: fmt.Sprintf("security group rule %d not found", ruleID)}
}

func ParseRuleImportID(value string) (int64, int64, error) {
	for _, sep := range []string{":", "/"} {
		for i, char := range value {
			if string(char) != sep {
				continue
			}
			sgID, err := strconv.ParseInt(value[:i], 10, 64)
			if err != nil {
				return 0, 0, err
			}
			ruleID, err := strconv.ParseInt(value[i+1:], 10, 64)
			if err != nil {
				return 0, 0, err
			}
			return sgID, ruleID, nil
		}
	}
	return 0, 0, fmt.Errorf("expected import ID in <security_group_id>:<rule_id> format")
}

func mapRulePayload(input SecurityGroupRuleInput, update bool) map[string]any {
	rule := map[string]any{}
	if !update {
		rule["ruleType"] = "customRule"
	}
	if input.Name != "" {
		rule["name"] = input.Name
	}
	if input.Direction != "" {
		rule["direction"] = input.Direction
	}
	if input.Policy != "" {
		rule["policy"] = input.Policy
	}
	if input.Protocol != "" {
		rule["protocol"] = input.Protocol
	}
	if input.PortRange != "" {
		rule["portRange"] = input.PortRange
	}
	if input.DestinationPortRange != "" {
		rule["destinationPortRange"] = input.DestinationPortRange
	}
	if input.SourceType != "" {
		rule["sourceType"] = input.SourceType
	}
	if input.Source != "" {
		rule["source"] = input.Source
	}
	if input.DestinationType != "" {
		rule["destinationType"] = input.DestinationType
	}
	if input.Destination != "" {
		rule["destination"] = input.Destination
	}
	if input.Ethertype != "" {
		rule["ethertype"] = input.Ethertype
	}
	if input.Priority != nil {
		rule["priority"] = *input.Priority
	}
	if input.Enabled != nil {
		rule["enabled"] = *input.Enabled
	}
	return map[string]any{"rule": rule}
}
