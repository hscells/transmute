package ir

// Terms extracts a list of query terms from the Boolean query.
func (b BooleanQuery) Terms() (s []string) {
	for _, keyword := range b.Keywords{
		s = append(s, keyword.QueryString)
	}
	for _, child := range b.Children {
		s = append(s, child.Terms()...)
	}
	return
}

// Fields extracts the fields from the query.
func (b BooleanQuery) Fields() (f []string) {
	for _, keyword := range b.Keywords{
		f = append(f, keyword.Fields...)
	}
	for _, child := range b.Children {
		f= append(f, child.Fields()...)
	}
	return
}

// FieldCount extracts the count of fields in a query.
func (b BooleanQuery) FieldCount() (c map[string]int) {
	c = map[string]int{}
	for _, field := range b.Fields() {
		c[field] += 1
	}
	return
}