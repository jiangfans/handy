package gorm_tools

import (
	"fmt"
	"gorm.io/gorm"
)

const DefaultPageSize = 20

type Direction int8

const (
	Desc Direction = 1
	Asc  Direction = 2
)

func (d Direction) string() string {
	if d == Desc {
		return "DESC"
	} else {
		return "ASC"
	}
}

type (
	Opts struct {
		ids            []uint64
		unscoped       bool
		offset, limit  int
		orderBy        string
		whereStatement string
		args           []interface{}
	}

	funcOption struct {
		f func(opts *Opts)
	}

	Option interface {
		apply(opts *Opts)
	}
)

func (fdo *funcOption) apply(do *Opts) {
	fdo.f(do)
}

func newFuncQueryOption(f func(opts *Opts)) *funcOption {
	return &funcOption{
		f: f,
	}
}

func Ids(ids []uint64) Option {
	return newFuncQueryOption(func(opts *Opts) {
		opts.ids = ids
	})
}

func Pagination(page, pageSize int32) Option {
	return newFuncQueryOption(func(opts *Opts) {
		if page == 0 {
			page = 1
		}
		if pageSize == 0 {
			pageSize = DefaultPageSize
		}

		opts.limit = int(pageSize)
		opts.offset = int(pageSize * (page - 1))
	})
}

func Limit(limit int32) Option {
	return newFuncQueryOption(func(opts *Opts) {
		opts.limit = int(limit)
	})
}

func OrderBy(column string, directions ...Direction) Option {
	direction := Asc
	if len(directions) != 0 {
		direction = directions[0]
	}

	return newFuncQueryOption(func(opts *Opts) {
		if opts.orderBy != "" {
			opts.orderBy += ","
		}

		opts.orderBy += column + " " + direction.string()
	})
}

func Equal(column string, value interface{}) Option {
	return newFuncQueryOption(func(opts *Opts) {
		if opts.whereStatement != "" {
			opts.whereStatement += " AND "
		}

		opts.whereStatement += fmt.Sprintf("%s = ?", column)
		opts.args = append(opts.args, value)
	})
}

func NotEqual(column string, value interface{}) Option {
	return newFuncQueryOption(func(opts *Opts) {
		if opts.whereStatement != "" {
			opts.whereStatement += " AND "
		}

		opts.whereStatement += fmt.Sprintf("%s != ?", column)
		opts.args = append(opts.args, value)
	})
}

func Gt(column string, value interface{}) Option {
	return newFuncQueryOption(func(opts *Opts) {
		if opts.whereStatement != "" {
			opts.whereStatement += " AND "
		}

		opts.whereStatement += fmt.Sprintf("%s > ?", column)
		opts.args = append(opts.args, value)
	})
}

func Gte(column string, value interface{}) Option {
	return newFuncQueryOption(func(opts *Opts) {
		if opts.whereStatement != "" {
			opts.whereStatement += " AND "
		}

		opts.whereStatement += fmt.Sprintf("%s >= ?", column)
		opts.args = append(opts.args, value)
	})
}

func Lt(column string, value interface{}) Option {
	return newFuncQueryOption(func(opts *Opts) {
		if opts.whereStatement != "" {
			opts.whereStatement += " AND "
		}

		opts.whereStatement += fmt.Sprintf("%s < ?", column)
		opts.args = append(opts.args, value)
	})
}

func Lte(column string, value interface{}) Option {
	return newFuncQueryOption(func(opts *Opts) {
		if opts.whereStatement != "" {
			opts.whereStatement += " AND "
		}

		opts.whereStatement += fmt.Sprintf("%s <= ?", column)
		opts.args = append(opts.args, value)
	})
}

func In(column string, array interface{}) Option {
	return newFuncQueryOption(func(opts *Opts) {
		if opts.whereStatement != "" {
			opts.whereStatement += " AND "
		}

		opts.whereStatement += fmt.Sprintf("%s in (?)", column)
		opts.args = append(opts.args, array)
	})
}

func Unscoped() Option {
	return newFuncQueryOption(func(opts *Opts) {
		opts.unscoped = true
	})
}

func DB(db *gorm.DB, optList []Option) *gorm.DB {
	opts := &Opts{}
	for _, opt := range optList {
		opt.apply(opts)
	}

	if len(opts.ids) != 0 {
		db = db.Where("id in (?)", opts.ids)
	}

	if opts.limit != 0 {
		db = db.Limit(opts.limit)
	}

	if opts.offset != 0 {
		db = db.Offset(opts.offset)
	}

	if opts.orderBy != "" {
		db = db.Order(opts.orderBy)
	} else {
		db = db.Order("id")
	}

	if opts.whereStatement != "" {
		db = db.Where(opts.whereStatement, opts.args...)
	}

	return db
}
