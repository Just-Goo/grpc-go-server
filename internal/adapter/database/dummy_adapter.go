package database

import (
	"log"

	"github.com/google/uuid"
)

func (d *DatabaseAdapter) Save(data *DummyOrm) (uuid.UUID, error) {
	if err := d.db.Create(data).Error; err != nil {
		log.Println("can't create data:", err)
		return uuid.Nil, err
	}

	return data.UserID, nil
}

func (d *DatabaseAdapter) GetByUUID(uuid *uuid.UUID) (DummyOrm, error) {
	var res DummyOrm
	if err := d.db.First(&res, "user_id = ?", uuid); err != nil {
		log.Println("can't create data:", err)
		return res, err.Error
	}

	return res, nil
}
