package commands

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

// setupAuthCommand создает команды аутентификации
func setupAuthCommand() *cobra.Command {
	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Команды аутентификации",
		Long:  `Регистрация новых пользователей и аутентификация существующих.`,
	}

	authCmd.AddCommand(
		setupRegisterCommand(),
		setupLoginCommand(),
	)

	return authCmd
}

// setupRegisterCommand создает команду регистрации
func setupRegisterCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "register",
		Short: "Зарегистрировать нового пользователя",
		Long:  `Регистрация нового пользователя в системе GophKeeper.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			login, _ := cmd.Flags().GetString("login")
			password, _ := cmd.Flags().GetString("password")

			ctx := context.Background()
			if err := app.Register(ctx, login, password); err != nil {
				return fmt.Errorf("ошибка регистрации: %w", err)
			}

			fmt.Println("Пользователь успешно зарегистрирован")
			return nil
		},
	}
}

// setupLoginCommand создает команду входа
func setupLoginCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Войти в систему",
		Long:  `Аутентификация в системе GophKeeper.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			login, _ := cmd.Flags().GetString("login")
			password, _ := cmd.Flags().GetString("password")

			ctx := context.Background()
			if err := app.Login(ctx, login, password); err != nil {
				return fmt.Errorf("ошибка входа: %w", err)
			}

			fmt.Println("Вход выполнен успешно")
			return nil
		},
	}
}

func init() {
	registerCmd := setupRegisterCommand()
	registerCmd.Flags().StringP("login", "l", "", "логин пользователя")
	registerCmd.Flags().StringP("password", "p", "", "пароль пользователя")
	registerCmd.MarkFlagRequired("login")
	registerCmd.MarkFlagRequired("password")

	loginCmd := setupLoginCommand()
	loginCmd.Flags().StringP("login", "l", "", "логин пользователя")
	loginCmd.Flags().StringP("password", "p", "", "пароль пользователя")
	loginCmd.MarkFlagRequired("login")
	loginCmd.MarkFlagRequired("password")
}
