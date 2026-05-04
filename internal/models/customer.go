package models

type Customer struct {
	CreatedAt			uint	`json:"createdAt"`
	UpdatedAt			uint	`json:"updatedAt"`
	Id						uint	`json:"id"`
	Archived			uint	`json:"archived"`
	Email					string	`json:"email"`
	EmailConfirmed	uint		`json:"email_confirmed"`
	FirstName	string	`json:"first_name"`
	LastNme	string	`json:"last_name"`
	Address1	string	`json:"address_1"`
	Address2	string	`json:"address_2"`
	ZipCode	string	`json:"zip_code"`
	City	string	`json:"city"`
	Country	string	`json:"country"`
	Phone	string	`json:"phone"`
	VatId	string	`json:"vat_id"`
	CompanyName	string	`json:"company_name"`
	Customer	uint		`json:"customer"`
	Teacher	uint		`json:"teacher"`
	Admin	uint		`json:"admin"`
	Owner	uint		`json:"owner"`
	Checkin	uint		`json:"checkin"`
	DateOfBirth		string	`json:"date_of_birth"`
	TeacherDescription	string	`json:"teacher_description"`
	CustomerAdditionalInfo	string	`json:"customer_additional_info"`
	AdminNotes	string	`json:"admin_notes"`
	ImportWelcomeSetPasswordEmailSent	uint `json:"import_welcome_set_password_email_sent"`
	TeacherCanManageAllClasses	uint `json:"teacher_can_manage_all_classes"`
	LivestreamTimeDisplayMode	string `json:"livestream_time_display_mode"`
	Newsletter	uint	`json:"newsletter"`
	KisiUserId	string	`json:"kisi_user_id"`
	Client	uint	`json:"client"`
	ForbidPurchases	uint	`json:"forbid_purchases"`
	ImageId	uint	`json:"image_id"`
}
