package types

type UserStatus string
type UserRole string
type UserType string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"
	UserStatusFrozen   UserStatus = "frozen"
	UserStatusDisabled UserStatus = "disabled"
	UserStatusLocked   UserStatus = "locked"
	UserStatusDeleted  UserStatus = "deleted"
	UserStatusMerged   UserStatus = "merged"
)

const (
	UserRoleAdmin UserRole = "admin"
	UserRoleUser  UserRole = "user"
	UserRoleGuest UserRole = "guest"
)

const (
	UserTypeCustomer UserType = "customer"
	UserTypeStaff    UserType = "staff"
	UserTypePartner  UserType = "partner"
)
