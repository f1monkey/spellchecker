package dictionary

import (
	"math"
	"sort"

	"github.com/agnivade/levenshtein"
)

type Match struct {
	Value string
	Score float64
}

func (d *Dictionary) Find(word string, n int, maxErrors int) []Match {
	d.mtx.RLock()
	defer d.mtx.RUnlock()

	if maxErrors <= 0 {
		return nil
	}

	bm := d.alphabet.encode([]rune(word))
	candidates := d.getCandidates(word, bm, 1, maxErrors)
	result := calcScores([]rune(word), candidates)

	if len(result) < n {
		return result
	}

	return result[0:n]
}

type Candidate struct {
	Word     string
	Distance int
	Count    int
}

func (d *Dictionary) getCandidates(word string, bmSrc bitmap, errCnt int, maxErrors int) []Candidate {
	checked := make(map[bitmap]struct{}, d.alphabet.len()*2)

	// exact match OR candidate has all the same letters as the word but in different order
	result := make([]Candidate, 0, 50)
	if _, ok := checked[bmSrc]; !ok {
		checked[bmSrc] = struct{}{}
		ids := d.index.get(bmSrc)
		for _, id := range ids {
			doc, ok := d.docRaw(id)
			if !ok {
				continue
			}

			distance := levenshtein.ComputeDistance(word, doc.Word)
			if distance > maxErrors {
				continue
			}
			result = append(result, Candidate{
				Word:     doc.Word,
				Count:    doc.Count,
				Distance: distance,
			})
		}
	}

	for i := 0; i < len(d.alphabet); i++ {
		bmCandidate := bmSrc.clone()
		bmCandidate.xor(uint32(i))
		if _, ok := checked[bmCandidate]; ok {
			continue
		}
		checked[bmCandidate] = struct{}{}

		ids := d.index.get(bmCandidate)
		for _, id := range ids {
			doc, ok := d.docRaw(id)
			if !ok {
				continue
			}

			distance := levenshtein.ComputeDistance(word, doc.Word)
			if distance > maxErrors {
				continue
			}
			result = append(result, Candidate{
				Word:     doc.Word,
				Count:    doc.Count,
				Distance: distance,
			})
		}

		for j := 0; j < len(d.alphabet); j++ {
			bmCandidate2 := bmCandidate.clone()
			bmCandidate2.xor(uint32(j))
			if _, ok := checked[bmCandidate2]; ok {
				continue
			}
			checked[bmCandidate2] = struct{}{}

			ids := d.index.get(bmCandidate2)
			for _, id := range ids {
				doc, ok := d.docRaw(id)
				if !ok {
					continue
				}

				distance := levenshtein.ComputeDistance(word, doc.Word)
				if distance > maxErrors {
					continue
				}
				result = append(result, Candidate{
					Word:     doc.Word,
					Count:    doc.Count,
					Distance: distance,
				})
			}
		}
	}

	return result
}

func calcScores(src []rune, candidates []Candidate) []Match {
	result := make([]Match, len(candidates))
	for i, c := range candidates {
		result[i] = Match{
			Value: c.Word,
			Score: calcScore(src, []rune(c.Word), c.Distance, c.Count),
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Score > result[j].Score })

	return result
}

func calcScore(src []rune, candidate []rune, distance int, cnt int) float64 {
	mult := math.Log1p(float64(cnt))
	// if first letters are the same, increase score
	if src[0] == candidate[0] {
		mult *= 1.5
		// if second letters are the same too, increase score even more
		if len(src) > 1 && len(candidate) > 1 && src[1] == candidate[1] {
			mult *= 1.5
		}
	}

	return 1 / (1 + float64(distance*distance)) * mult
}

func abs(x int) int {
	if x < 0 {
		return -1 * x
	}
	return x
}
