package plugin

import "testing"

func TestPipelineContract(t *testing.T) {
 valid := PipelineSpec{Steps: []PipelineStep{{Helper:"string.trim",Input:"$input.name",Save:"name"}},Return:"$steps.name"}
 if !valid.valid() { t.Fatal("valid pipeline rejected") }
 for _, step := range []PipelineStep{
  {Helper:"exec",Input:"$input",Save:"name"},
  {Helper:"value.identity",Input:"$steps.future",Save:"name"},
  {Helper:"value.identity",Input:"$input.__proto__",Save:"name"},
  {Helper:"value.identity",Input:"$input",Save:"constructor"},
 } {
  invalid := PipelineSpec{Steps: []PipelineStep{step},Return:"$steps.name"}
  if invalid.valid() { t.Errorf("invalid pipeline accepted: %+v",step) }
 }
}
