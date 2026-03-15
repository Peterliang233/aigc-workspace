package settings

import (
	"context"
	"time"

	"gorm.io/gorm"
)

func (st *MySQLStore) Update(fn func(*Settings) error) (Settings, error) {
	st.mu.Lock()
	defer st.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var out Settings
	err := st.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		cur, err := st.getTx(tx)
		if err != nil {
			return err
		}
		next := deepCopy(cur)
		if err := fn(&next); err != nil {
			return err
		}
		if err := st.putTx(tx, next); err != nil {
			return err
		}
		out = next
		return nil
	})
	if err != nil {
		return Settings{}, err
	}
	return out, nil
}

