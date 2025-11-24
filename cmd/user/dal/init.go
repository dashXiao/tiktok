package dal

import (
	"github.com/ozline/tiktok/cmd/user/dal/cache"
	"github.com/ozline/tiktok/cmd/user/dal/db"
)

func Init() {
	db.Init()
	cache.Init()
}
