// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

const (
	Map = iota + 1
	Array
	Value

	Fail     = iota + 1 // Fail if a conflict occurs.
	Latest              // Latest configuration is kept if a conflict occurs.
	Existing            // Existing configuration is kept if a conflict occurs.
	Append              // Append the new values to existing values if a conflict occurs.
	Prepend             // Prepend the new values to existing values if a conflict occurs.
)

// OnLeafConflict option defines the behavior when merging configuration files
// and a leaf has a collision with another leaf.  A leaf node is a common data
// that represents a specific configuration value.  [Fail] An example of a leaf node:
//
//    ---
//    foo:
//        bar: 'bananas'    # bar & bananas represent a leaf node
//
// Options available are: [Fail] [Latest] [Existing]
func WithMergeStrategy(kind, strategy int) Option {
	return func(g *Goschtalt) error {
		switch kind {
		case Value:
			switch strategy {
			case Fail:
				g.leafConflictFn = func(cur, next annotatedValue) (annotatedValue, error) {
					return annotatedValue{}, ErrConflict
				}
			case Latest:
				g.leafConflictFn = func(cur, next annotatedValue) (annotatedValue, error) {
					return next, nil
				}
			case Existing:
				g.leafConflictFn = func(cur, next annotatedValue) (annotatedValue, error) {
					return cur, nil
				}
			default:
				return ErrInvalidOption
			}
		case Array:
			switch strategy {
			case Fail:
				g.arrayConflictFn = func(cur, next annotatedArray) (annotatedArray, error) {
					return annotatedArray{}, ErrConflict
				}
			case Latest:
				g.arrayConflictFn = func(cur, next annotatedArray) (annotatedArray, error) {
					return next, nil
				}
			case Existing:
				g.arrayConflictFn = func(cur, next annotatedArray) (annotatedArray, error) {
					return cur, nil
				}
			case Append:
				g.arrayConflictFn = func(cur, next annotatedArray) (annotatedArray, error) {
					cur.files = dedupedAppend(cur.files, next.files...)
					cur.array = append(cur.array, next.array...)
					return cur, nil
				}
			case Prepend:
				g.arrayConflictFn = func(cur, next annotatedArray) (annotatedArray, error) {
					cur.files = dedupedAppend(cur.files, next.files...)
					cur.array = append(next.array, cur.array...)
					return cur, nil
				}
			default:
				return ErrInvalidOption
			}
		case Map:
			switch strategy {
			case Fail:
				g.mapConflictFn = func(cur, next any) (any, error) {
					return annotatedMap{}, ErrConflict
				}
			case Latest:
				g.mapConflictFn = func(cur, next any) (any, error) {
					return next, nil
				}
			case Existing:
				g.mapConflictFn = func(cur, next any) (any, error) {
					return cur, nil
				}
			default:
				return ErrInvalidOption
			}
		default:
			return ErrInvalidOption
		}
		return nil
	}
}
