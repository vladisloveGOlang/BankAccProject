package domain

type Like struct {
	CreatedAt int64  `json:"created_at"`
	CreatedBy string `json:"created_by"`
}

type UserLike struct {
	CreatedAt int64 `json:"created_at"`
	User      User  `json:"user"`
}
