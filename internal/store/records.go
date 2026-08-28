package store

import (
	"go.etcd.io/bbolt"
	"sort"
	"tastinginvite/internal/codec"
	"tastinginvite/internal/model"
)

func (s *Store) PutRecord(record model.Record) error {
	data, err := codec.EncodeRecord(record)
	if err != nil {
		return err
	}
	return s.Update(func(tx *bbolt.Tx) error { return tx.Bucket([]byte("records")).Put([]byte(record.ID), data) })
}

func (s *Store) GetRecord(id string) (model.Record, error) {
	var result model.Record
	err := s.View(func(tx *bbolt.Tx) error {
		value := tx.Bucket([]byte("records")).Get([]byte(id))
		if value == nil {
			return ErrNotFound
		}
		var err error
		result, err = codec.DecodeRecord(cloneBytes(value))
		return err
	})
	return result, err
}

func (s *Store) DeleteRecord(id string) error {
	return s.Update(func(tx *bbolt.Tx) error { return tx.Bucket([]byte("records")).Delete([]byte(id)) })
}

func (s *Store) ListRecords() ([]model.Record, error) {
	records := make([]model.Record, 0)
	err := s.View(func(tx *bbolt.Tx) error {
		return tx.Bucket([]byte("records")).ForEach(func(_, value []byte) error {
			if value == nil {
				return nil
			}
			r, err := codec.DecodeRecord(cloneBytes(value))
			if err != nil {
				return err
			}
			records = append(records, r)
			return nil
		})
	})
	sort.Slice(records, func(i, j int) bool { return records[i].ID < records[j].ID })
	return records, err
}

func (s *Store) PutAudit(event model.AuditEvent) error {
	data, err := codec.EncodeAudit(event)
	if err != nil {
		return err
	}
	return s.Update(func(tx *bbolt.Tx) error { return tx.Bucket([]byte("audits")).Put([]byte(event.ID), data) })
}

func (s *Store) ListAudits(recordID string) ([]model.AuditEvent, error) {
	out := make([]model.AuditEvent, 0)
	err := s.View(func(tx *bbolt.Tx) error {
		return tx.Bucket([]byte("audits")).ForEach(func(_, value []byte) error {
			if value == nil {
				return nil
			}
			event, e := codec.DecodeAudit(cloneBytes(value))
			if e != nil {
				return e
			}
			if event.RecordID == recordID {
				out = append(out, event)
			}
			return nil
		})
	})
	sort.Slice(out, func(i, j int) bool { return out[i].At.Before(out[j].At) })
	return out, err
}
