package audited

import "fmt"

type AuditedModel struct {
	CreatedBy string
	UpdatedBy string
}

// SetCreatedBy set created by
func (model *AuditedModel) SetCreatedBy(createdBy interface{}) {
	model.CreatedBy = fmt.Sprintf("%v", createdBy)
}

// GetCreatedBy get created by
func (model AuditedModel) GetCreatedBy() string {
	return model.CreatedBy
}

// SetUpdatedBy set updated by
func (model *AuditedModel) SetUpdatedBy(updatedBy interface{}) {
	model.UpdatedBy = fmt.Sprintf("%v", updatedBy)
}

// GetUpdatedBy get updated by
func (model AuditedModel) GetUpdatedBy() string {
	return model.UpdatedBy
}
