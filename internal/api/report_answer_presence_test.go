package api

import (
	"encoding/json"
	pb "github.com/LimeOnTop/interverse-contracts/report/gen"
	"google.golang.org/protobuf/proto"
	"testing"
)

func TestReportAnswerPresenceSurvivesWire(t *testing.T) {
	var input sessionAnswersRequest
	if err := json.Unmarshal([]byte(`[{"step_id":"skipped","item_type":"question"},{"step_id":"first","item_type":"question","selected_option_index":0},{"step_id":"second","item_type":"question","selected_option_index":1}]`), &input); err != nil {
		t.Fatal(err)
	}
	data, err := proto.Marshal(&pb.GenerateReportRequest{Answers: toSessionAnswerInputs(input)})
	if err != nil {
		t.Fatal(err)
	}
	var decoded pb.GenerateReportRequest
	if err = proto.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Answers[0].SelectedOptionIndex != nil {
		t.Fatal("a skipped answer became option A")
	}
	if decoded.Answers[1].SelectedOptionIndex == nil || decoded.Answers[1].GetSelectedOptionIndex() != 0 {
		t.Fatal("explicit option A was lost")
	}
	if decoded.Answers[2].GetSelectedOptionIndex() != 1 {
		t.Fatal("explicit option B was changed")
	}
}
