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

	checked := make(map[bitmap]struct{})
	bm := d.alphabet.encode([]rune(word))
	candidates := d.getCandidates(word, bm, 1, maxErrors, checked)
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

func (d *Dictionary) getCandidates(word string, bmSrc bitmap, errCnt int, maxErrors int, checked map[bitmap]struct{}) []Candidate {
	// exact match OR candidate has all the same letters as the word but in different order
	result := make([]Candidate, 0, len(d.alphabet))
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

	return append(result, d.getCandidatesRecursive(word, bmSrc, 1, 2, checked)...)
}

func (d *Dictionary) getCandidatesRecursive(word string, bmSrc bitmap, errCnt int, maxErrors int, checked map[bitmap]struct{}) []Candidate {
	if errCnt > maxErrors {
		return nil
	}

	result := make([]Candidate, 0, len(d.alphabet))
	for i := 0; i < len(d.alphabet); i++ {
		bmCandidate := bmSrc.clone()
		bmCandidate.xor(uint32(i))

		if _, ok := checked[bmCandidate]; ok {
			continue
		}
		checked[bmCandidate] = struct{}{}

		diff := bmSrc.countDiff(bmCandidate)
		if diff > maxErrors {
			continue
		}

		ids := d.index[bmCandidate]
		if len(ids) != 0 {
			for _, id := range ids {
				doc, ok := d.docRaw(id)
				if !ok {
					continue
				}

				distance := levenshtein.ComputeDistance(word, doc.Word)
				if distance+diff > maxErrors {
					continue
				}
				result = append(result, Candidate{
					Word:     doc.Word,
					Count:    doc.Count,
					Distance: distance,
				})
			}
		}

		result = append(result, d.getCandidatesRecursive(word, bmCandidate, errCnt+1, maxErrors, checked)...)
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
