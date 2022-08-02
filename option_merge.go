// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

// MergeStrategy inputs
const (
	// MergeStrategy kind values
	Map = iota + 1
	Array
	Value

	// MergeStrategy strategy values
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

// MergeStrategy defines how different object types are merged together.
// Each kind of object (Map, Array, Value) has a definition of how it is merged.
//
//   - Array, Append - Arrays are appended to existing arrays in file priorty order
//     (lowest -> highest) so the highest prioritized items are at the bottom.
//   - Array, Existing - Newly processed arrays are dropped in favor of existing.
//   - Array, Fail - Duplicate arrays triger a failure in processing.
//   - Array, Latest - Arrays are replaced by newly processed arrays.
//   - Array, Prepend - Arrays are prepended to existing arrays in file priorty
//     order (highest -> lowest) so the highest prioritized items are at the top.
//   - Map, Existing - Newly processed maps are dropped in favor of existing.
//   - Map, Fail - Duplicate maps triger a failure in processing.
//   - Map, Latest - Maps are replaced by newly processed arrays.
//   - Value, Existing - Newly processed values are dropped in favor of existing.
//   - Value, Fail - Duplicate values triger a failure in processing.
//   - Value, Latest - Values are replaced by newly processed arrays.
func MergeStrategy(kind, strategy int) Option {
	return func(g *Goschtalt) error {
		switch kind {
		case Value:
			switch strategy {
			case Fail:
				g.valueConflictFn = func(cur, next annotatedValue) (annotatedValue, error) {
					return annotatedValue{}, ErrConflict
				}
			case Latest:
				g.valueConflictFn = func(cur, next annotatedValue) (annotatedValue, error) {
					return next, nil
				}
			case Existing:
				g.valueConflictFn = func(cur, next annotatedValue) (annotatedValue, error) {
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
