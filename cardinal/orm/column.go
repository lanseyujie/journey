package orm

type Column struct {
    Name    string // for model
    Alias   string // for sql
    Type    string
    Options string
}

// GetAlias
func (col *Column) GetAlias() string {
    if col.Alias != "" {
        return col.Alias
    }

    return UnderScoreCase(col.Name)
}

// UnderScoreCase convert camel case to under score case
func UnderScoreCase(camel string) string {
    cc := []byte(camel)
    usc := make([]byte, 0, len(cc))
    for index, ascii := range cc {
        if 'A' <= ascii && ascii <= 'Z' {
            if index > 0 {
                usc = append(usc, '_')
            }
            // convert to lower case
            // ASCII A~Z => 65~90 a~z => 97~122
            usc = append(usc, byte(int(ascii)+97-65))
        } else {
            usc = append(usc, ascii)
        }
    }

    return string(usc)
}
