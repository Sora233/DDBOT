package blockCache

const offset32 = 2166136261
const prime32 = 16777619

type sum32 uint32

func (s *sum32) Reset()         { *s = offset32 }
func (s *sum32) Size() int      { return 4 }
func (s *sum32) BlockSize() int { return 1 }
func (s *sum32) Write(data []byte) (int, error) {
	hash := *s
	for _, c := range data {
		hash *= prime32
		hash ^= sum32(c)
	}
	*s = hash
	return len(data), nil
}
func (s *sum32) Sum32() uint32 { return uint32(*s) }

func fnvHasher(b []byte) uint32 {
	s := new(sum32)
	s.Write(b)
	return s.Sum32()
}
