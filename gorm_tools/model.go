package gorm_tools

import (
	"reflect"
	"time"

	"github.com/bxcodec/faker/v3"
	xid "gitlab.shoplazza.site/common/common-xid"
	"gorm.io/gorm"
)

var _ = faker.AddProvider("primary_id", func(v reflect.Value) (interface{}, error) {
	value := xid.Get()
	if v.Kind() == reflect.Ptr {
		return value, nil
	}
	return value, nil
})

type BaseModel struct {
	// id
	Id uint64 `gorm:"column:id" faker:"primary_id" json:"id" example:"94344029373746188"`
	// 创建时间
	CreatedAt *Time `gorm:"column:created_at" faker:"-" json:"created_at" example:"2022-10-24T00:00:00Z"`
	// 更新时间
	UpdatedAt *Time          `gorm:"column:updated_at" faker:"-" json:"updated_at" example:"2022-10-24T00:00:00Z"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at" faker:"-" json:"-"`
}

func (b *BaseModel) BeforeCreate(tx *gorm.DB) (err error) {
	now := time.Now()
	if b.Id == 0 {
		tx.Statement.SetColumn("id", xid.Get())
	}

	createdAt, updatedAt := Time(now), Time(now)
	tx.Statement.SetColumn("created_at", &createdAt)
	tx.Statement.SetColumn("updated_at", &updatedAt)
	return
}

func (b *BaseModel) AfterUpdate(tx *gorm.DB) (err error) {
	tx.Statement.SetColumn("updated_at", Time(time.Now()))
	return
}

// ________

type CreatedAt struct {
	// 创建时间
	CreatedAt Time `gorm:"column:created_at" faker:"-" json:"created_at" example:"2021-11-01T00:00:00Z"`
}

func (c *CreatedAt) BeforeCreate(tx *gorm.DB) (err error) {
	tx.Statement.SetColumn("created_at", Time(time.Now()))
	return
}
