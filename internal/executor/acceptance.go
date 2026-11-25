package executor

import (
	"context"
	"encoding/json"
	"errors"

	aw "github.com/yesz1520/autowonder-switch/internal/autowonder"
	"github.com/yesz1520/autowonder-switch/internal/harness"
)

const acceptanceSchema = `{
  "name":"executor.report_acceptance",
  "description":"Report acceptance for requirements in this work item's frozen design/requirements.json. Use its exact keys. PASSED requires evidence. NOT_APPLICABLE requires a reason.",
  "inputSchema":{"type":"object","additionalProperties":false,"required":["reports"],"properties":{
    "reports":{"type":"array","minItems":1,"items":{"type":"object","additionalProperties":false,
      "required":["requirementKey","implementation","verdict"],"properties":{
        "requirementKey":{"type":"string"},
        "implementation":{"type":"string","enum":["IMPLEMENTED","PARTIAL","NOT_IMPLEMENTED"]},
        "verdict":{"type":"string","enum":["PASSED","FAILED","NOT_VERIFIED","NOT_APPLICABLE"]},
        "note":{"type":"string"},"notApplicableReason":{"type":"string"},
        "evidence":{"type":"array","items":{"type":"object","additionalProperties":false,"required":["kind","ref"],"properties":{
          "kind":{"type":"string","enum":["ARTIFACT","CODE","MR","TEST","PREVIEW","URL"]},"ref":{"type":"string"},"title":{"type":"string"}
        }}}
      }
    }}
  }}
}`

func (r *runtime) withAcceptance(platform harness.PlatformTool, dispatchID int64) harness.PlatformTool {
	return func(ctx context.Context, name string, args map[string]any) (string, error) {
		switch name {
		case "":
			return appendTool(ctx, platform, json.RawMessage(acceptanceSchema))
		case "executor.report_acceptance":
			data, err := json.Marshal(args["reports"])
			if err != nil {
				return "", err
			}
			var reports []aw.AcceptanceReport
			if err := json.Unmarshal(data, &reports); err != nil {
				return "", err
			}
			if len(reports) == 0 {
				return "", errors.New("at least one requirement report is required")
			}
			receipt, err := r.client.ReportAcceptance(ctx, dispatchID, reports)
			if err != nil {
				return "", err
			}
			data, err = json.Marshal(receipt)
			return string(data), err
		default:
			return platform(ctx, name, args)
		}
	}
}
