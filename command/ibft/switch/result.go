package ibftswitch

import (
	"bytes"
	"fmt"

	"github.com/dogechain-lab/dogechain/command/helper"
	"github.com/dogechain-lab/dogechain/consensus/ibft"
	"github.com/dogechain-lab/dogechain/helper/common"
)

type IBFTSwitchResult struct {
	Chain      string             `json:"chain"`
	Type       ibft.MechanismType `json:"type"`
	From       common.JSONNumber  `json:"from"`
	Deployment *common.JSONNumber `json:"deployment,omitempty"`
}

func (r *IBFTSwitchResult) GetOutput() string {
	var buffer bytes.Buffer

	buffer.WriteString("\n[NEW IBFT FORK]\n")

	outputs := []string{
		fmt.Sprintf("Chain|%s", r.Chain),
		fmt.Sprintf("Type|%s", r.Type),
	}
	if r.Deployment != nil {
		outputs = append(outputs, fmt.Sprintf("Deployment|%d", r.Deployment.Value))
	}

	outputs = append(outputs, fmt.Sprintf("From|%d", r.From.Value))

	buffer.WriteString(helper.FormatKV(outputs))
	buffer.WriteString("\n")

	return buffer.String()
}
