package ds

import "time"

type LoadDTO struct {
	ID                     uint    `json:"id"`
	LoadTitle              string  `json:"load_title"`
	LoadDescription        string  `json:"load_description"`
	LoadImage              *string `json:"load_image"`
	Normative              float64 `json:"normative"`
	LoadCategory           string  `json:"load_category"`
	ReliabilityCoefficient float64 `json:"reliability_coefficient"`
	Status                 *bool   `json:"status"`
}

type LoadCreateRequest struct {
	LoadTitle              string  `json:"load_title" binding:"required"`
	LoadDescription        string  `json:"load_description" binding:"required"`
	Normative              float64 `json:"normative" binding:"required"`
	LoadCategory           string  `json:"load_category" binding:"required"`
	ReliabilityCoefficient float64 `json:"reliability_coefficient" binding:"required"`
}

type LoadUpdateRequest struct {
	LoadTitle              *string  `json:"load_title"`
	LoadDescription        *string  `json:"load_description"`
	Normative              *float64 `json:"normative"`
	LoadCategory           *string  `json:"load_category"`
	ReliabilityCoefficient *float64 `json:"reliability_coefficient"`
}

type LoadSessionDTO struct {
	ID          uint               `json:"id"`
	Status      int                `json:"status"`
	CreatedAt   time.Time          `json:"created_at"`
	CreatorID   uint               `json:"creator_id"`
	ModeratorID *uint              `json:"moderator_id"`
	RoomType    *string            `json:"room_type"`
	TotalLoad   *float64           `json:"total_load,omitempty"`
	Loads       []LoadInSessionDTO `json:"loads,omitempty"`
}

type LoadInSessionDTO struct {
	LoadID       uint    `json:"load_id"`
	LoadTitle    string  `json:"load_title"`
	LoadImage    *string `json:"load_image"`
	Normative    float64 `json:"normative"`
	LoadCategory string  `json:"load_category"`
	Area         *int    `json:"area"`
}

type LoadSessionUpdateRequest struct {
	RoomType *string `json:"room_type"`
}

type LoadSessionFormRequest struct {
}

type LoadSessionResolveRequest struct {
	Action string `json:"action" binding:"required"`
}

type LoadToCalculationUpdateRequest struct {
	Area *int `json:"area"`
}

type CartBadgeDTO struct {
	LoadSessionID *uint `json:"load_session_id"`
	LoadsCount    int   `json:"loads_count"`
}

type UserDTO struct {
	ID        uint   `json:"id"`
	Username  string `json:"username"`
	Moderator bool   `json:"moderator"`
}

type UserRegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserUpdateRequest struct {
	Username *string `json:"username"`
	Password *string `json:"password"`
}

type LoginResponse struct {
	Token string  `json:"token"`
	User  UserDTO `json:"user"`
}

type PaginatedResponse struct {
	Items interface{} `json:"items"`
	Total int64       `json:"total"`
}

type ErrorResponse struct {
	Status      string `json:"status"`
	Description string `json:"description"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}
