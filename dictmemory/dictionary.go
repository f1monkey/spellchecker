package dictmemory

type Dict struct {
	NextID uint32
	IDs    map[string]uint32
	Counts map[uint32]int
}

func NewDict() *Dict {
	return &Dict{
		NextID: 1,
		IDs:    make(map[string]uint32),
		Counts: make(map[uint32]int),
	}
}

func (d *Dict) ID(word string) (uint32, error) {
	return d.IDs[word], nil
}

func (d *Dict) Add(word string) (uint32, error) {
	if id, ok := d.IDs[word]; ok {
		d.Counts[id]++
		return id, nil
	}

	id := d.NextID
	d.IDs[word] = id
	d.Counts[id] = 1
	d.NextID++
	return id, nil
}
