package dictionary

import (
	"math"
	"sort"

	"github.com/agnivade/levenshtein"
)

// maxErrros is not a "max errors" in a word. It is a max diff in bits betweeen the "search word" and a "dictionary word".
// i.e. one simple symbol replacement (problam => problem ) is a two-bit difference.
const maxErrors = 2

type Match struct {
	Value string
	Score float64
}

func (d *Dictionary) Find(word string, n int) []Match {
	d.mtx.RLock()
	defer d.mtx.RUnlock()

	if maxErrors <= 0 {
		return nil
	}

	bm := d.alphabet.encode([]rune(word))
	candidates := d.getCandidates(word, bm, 1)
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

func (d *Dictionary) getCandidates(word string, bmSrc bitmap, errCnt int) []Candidate {
	checked := make(map[bitmap]struct{}, d.alphabet.len()*2)

	// "exact match" OR "candidate has all the same letters as the word but in different order"
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
	// the most common mistake is a transposition of letters.
	// so if we found one here, we do early termination
	if len(result) != 0 {
		return result
	}

	// @todo perform phonetic analysis with early termination here

	// @todo try to use tree index here
	toCheck := []bitmap{bmSrc}
	for errCnt := 1; errCnt <= maxErrors; errCnt++ {
		toCheck2 := make([]bitmap, 0, d.alphabet.len())
		for _, bm := range toCheck {
			for i := 0; i < len(d.alphabet); i++ {
				bmCandidate := bm.clone()
				bmCandidate.xor(uint32(i))
				if _, ok := checked[bmCandidate]; ok {
					continue
				}
				checked[bmCandidate] = struct{}{}
				toCheck2 = append(toCheck2, bmCandidate)

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
			}
		}
		toCheck = toCheck2
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
