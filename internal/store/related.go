package store

import (
	"go.etcd.io/bbolt"
	"tastinginvite/internal/codec"
	"tastinginvite/internal/model"
)

func (s *Store) PutWorkflow(workflow model.Workflow) error {
	data, err := codec.EncodeWorkflow(workflow)
	if err != nil {
		return err
	}
	return s.Update(func(tx *bbolt.Tx) error { return tx.Bucket([]byte("workflows")).Put([]byte(workflow.ID), data) })
}

func (s *Store) ListWorkflows(recordID string) ([]model.Workflow, error) {
	out := make([]model.Workflow, 0)
	err := s.View(func(tx *bbolt.Tx) error {
		return tx.Bucket([]byte("workflows")).ForEach(func(_, value []byte) error {
			if value == nil {
				return nil
			}
			w, e := codec.DecodeWorkflow(cloneBytes(value))
			if e != nil {
				return e
			}
			if w.RecordID == recordID {
				out = append(out, w)
			}
			return nil
		})
	})
	return out, err
}

func (s *Store) PutAttachment(attachment model.Attachment) error {
	data, err := codec.EncodeAttachment(attachment)
	if err != nil {
		return err
	}
	return s.Update(func(tx *bbolt.Tx) error { return tx.Bucket([]byte("attachments")).Put([]byte(attachment.ID), data) })
}

func (s *Store) ListAttachments(recordID string) ([]model.Attachment, error) {
	out := make([]model.Attachment, 0)
	err := s.View(func(tx *bbolt.Tx) error {
		return tx.Bucket([]byte("attachments")).ForEach(func(_, value []byte) error {
			if value == nil {
				return nil
			}
			a, e := codec.DecodeAttachment(cloneBytes(value))
			if e != nil {
				return e
			}
			if a.RecordID == recordID {
				out = append(out, a)
			}
			return nil
		})
	})
	return out, err
}
