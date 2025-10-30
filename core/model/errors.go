package model

import "errors"

// Validation errors for IdSearchReq

// ErrInvalidLevel is returned when a level value is negative.
var ErrInvalidLevel = errors.New("level must be non-negative")

// ErrLevelExceedsMax is returned when a level value is >= MaxLookupTableLevel.
var ErrLevelExceedsMax = errors.New("level exceeds maximum lookup table level")

// ErrInvalidDirection is returned when a direction value is neither DirectionLeft nor DirectionRight.
var ErrInvalidDirection = errors.New("direction must be either DirectionLeft or DirectionRight")

// Validation errors for Identifier

// ErrIdentifierTooLarge is returned when attempting to convert a byte slice larger than IdentifierSizeBytes to an Identifier.
var ErrIdentifierTooLarge = errors.New("input length exceeds identifier size")

// ErrInvalidHexString is returned when attempting to decode an invalid hex string to an Identifier.
var ErrInvalidHexString = errors.New("failed to decode hex string")

// Validation errors for MembershipVector

// ErrNegativeNumBits is returned when numBits parameter is negative.
var ErrNegativeNumBits = errors.New("numBits must be non-negative")

// ErrNumBitsExceedsMax is returned when numBits parameter exceeds the membership vector size.
var ErrNumBitsExceedsMax = errors.New("numBits exceeds membership vector size")

// ErrMembershipVectorTooLarge is returned when attempting to convert a byte slice larger than MembershipVectorSize to a MembershipVector.
var ErrMembershipVectorTooLarge = errors.New("input length exceeds membership vector size")
