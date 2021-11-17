package forms

type PassWordLoginForm struct {
	Mobile   string `form:"mobile" binding:"required"`
	PassWord string `form:"password" binding:"required,min=3,max=10"`
}
