package observex

func cloneFields(fields []Field) []Field {
	if len(fields) == 0 {
		return nil
	}
	copied := make([]Field, len(fields))
	copy(copied, fields)
	return copied
}

func redactFields(redactor Redactor, fields []Field) []Field {
	if redactor == nil {
		return cloneFields(fields)
	}
	return redactor.RedactFields(fields)
}
