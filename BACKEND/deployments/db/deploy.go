package deploy

import (
	"kvd/deployments/db/structure"
)

// Запускаем создание структуры БД и заполнение стартовой информацией о гильдии с АПИ raid.io через defer firstfilldb.FirstFillDB()
func init() {

}

func Deploy() {
	structure.Init()
}
