package util

import (
	"github.com/spf13/viper"
)

// Configurations wraps all the config variables required by the auth service
type Configurations struct {
	AccessTokenPrivateKeyPath  string
	AccessTokenPublicKeyPath   string
	RefreshTokenPrivateKeyPath string
	RefreshTokenPublicKeyPath  string
	JwtExpiration              int // in minutes
	SendGridApiKey             string
	MailVerifCodeExpiration    int // in hours
	PassResetCodeExpiration    int // in minutes
	MailVerifTemplateID        string
	PassResetTemplateID        string
}

// NewConfigurations returns a new Configuration object
func NewConfigurations() *Configurations {

	viper.AutomaticEnv()
	viper.SetDefault("ACCESS_TOKEN_PRIVATE_KEY_PATH", "./auth-private.pem")
	viper.SetDefault("ACCESS_TOKEN_PUBLIC_KEY_PATH", "./auth-public.pem")
	viper.SetDefault("REFRESH_TOKEN_PRIVATE_KEY_PATH", "./refresh-private.pem")
	viper.SetDefault("REFRESH_TOKEN_PUBLIC_KEY_PATH", "./refresh-public.pem")
	viper.SetDefault("JWT_EXPIRATION", 30)
	viper.SetDefault("MAIL_VERIFICATION_CODE_EXPIRATION", 24)
	viper.SetDefault("PASSWORD_RESET_CODE_EXPIRATION", 15)
	viper.SetDefault("MAIL_VERIFICATION_TEMPLATE_ID", "d-5ecbea6e38764af3b703daf03f139b48")
	viper.SetDefault("PASSWORD_RESET_TEMPLATE_ID", "d-3fc222d11809441abaa8ed459bb44319")

	configs := &Configurations{
		// ServerAddress:              viper.GetString("SERVER_ADDRESS"),
		// DBHost:                     viper.GetString("DB_HOST"),
		// DBName:                     viper.GetString("DB_NAME"),
		// DBUser:                     viper.GetString("DB_USER"),
		// DBPass:                     viper.GetString("DB_PASSWORD"),
		// DBPort:                     viper.GetString("DB_PORT"),
		// DBConn:                     conn,
		JwtExpiration:              viper.GetInt("JWT_EXPIRATION"),
		AccessTokenPrivateKeyPath:  viper.GetString("ACCESS_TOKEN_PRIVATE_KEY_PATH"),
		AccessTokenPublicKeyPath:   viper.GetString("ACCESS_TOKEN_PUBLIC_KEY_PATH"),
		RefreshTokenPrivateKeyPath: viper.GetString("REFRESH_TOKEN_PRIVATE_KEY_PATH"),
		RefreshTokenPublicKeyPath:  viper.GetString("REFRESH_TOKEN_PUBLIC_KEY_PATH"),
		SendGridApiKey:             viper.GetString("SENDGRID_API_KEY"),
		MailVerifCodeExpiration:    viper.GetInt("MAIL_VERIFICATION_CODE_EXPIRATION"),
		PassResetCodeExpiration:    viper.GetInt("PASSWORD_RESET_CODE_EXPIRATION"),
		MailVerifTemplateID:        viper.GetString("MAIL_VERIFICATION_TEMPLATE_ID"),
		PassResetTemplateID:        viper.GetString("PASSWORD_RESET_TEMPLATE_ID"),
	}

	// reading heroku provided port to handle deployment with heroku
	// port := viper.GetString("PORT")
	// if port != "" {
	// 	logger.Log().Debug("using the port allocated by heroku", port)
	// 	configs.ServerAddress = "0.0.0.0:" + port
	// }

	// logger.Log().Debug("serve port", configs.ServerAddress)
	// logger.Log().Debug("db host", configs.DBHost)
	// logger.Log().Debug("db name", configs.DBName)
	// logger.Log().Debug("db port", configs.DBPort)
	//logger.Log().Infof("jwt expiration", configs.JwtExpiration)

	return configs
}
