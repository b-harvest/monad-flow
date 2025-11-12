package state_sync

type StateSyncVersion struct {
	Major uint16
	Minor uint16
}

func (v StateSyncVersion) Ge(other StateSyncVersion) bool {
	vVal := (uint32(v.Major) << 16) | uint32(v.Minor)
	otherVal := (uint32(other.Major) << 16) | uint32(other.Minor)
	return vVal >= otherVal
}

func (v StateSyncVersion) Le(other StateSyncVersion) bool {
	vVal := (uint32(v.Major) << 16) | uint32(v.Minor)
	otherVal := (uint32(other.Major) << 16) | uint32(other.Minor)
	return vVal <= otherVal
}

var (
	STATESYNC_VERSION_V0  = StateSyncVersion{Major: 1, Minor: 0}
	STATESYNC_VERSION_V1  = StateSyncVersion{Major: 1, Minor: 1}
	STATESYNC_VERSION_V2  = StateSyncVersion{Major: 1, Minor: 2} // SELF_STATESYNC_VERSION
	STATESYNC_VERSION_MIN = STATESYNC_VERSION_V0
)

func (v StateSyncVersion) IsCompatible() bool {
	return v.Ge(STATESYNC_VERSION_MIN) && v.Le(STATESYNC_VERSION_V2)
}

type StateSyncBadVersion struct {
	MinVersion StateSyncVersion
	MaxVersion StateSyncVersion
}
