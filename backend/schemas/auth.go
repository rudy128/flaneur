package schemas

type SignupRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
	Email string `json:"email"`
}

type MessageResponse struct {
	Message string `json:"message"`
}
type TwitterAccountRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type GetTweetsRequest struct {
	URL string `json:"url" binding:"required"`
}

type TwitterLoginRequest struct {
	Username string `json:"username" binding:"required"`
}

type GetLikesRequest struct {
	URL string `json:"url" binding:"required"`
}

type GetQuotesRequest struct {
	URL string `json:"url" binding:"required"`
}

type GetCommentsRequest struct {
	URL string `json:"url" binding:"required"`
}

type GetRepostsRequest struct {
	URL string `json:"url" binding:"required"`
}