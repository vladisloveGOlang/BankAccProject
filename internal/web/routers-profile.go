package web

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/krisch/crm-backend/domain"
	"github.com/krisch/crm-backend/dto"
	"github.com/krisch/crm-backend/internal/emails"
	"github.com/krisch/crm-backend/internal/helpers"
	"github.com/krisch/crm-backend/internal/jwt"
	"github.com/krisch/crm-backend/internal/profile"
	oapi "github.com/krisch/crm-backend/internal/web/oprofile"
	echo "github.com/labstack/echo/v4"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

func initOpenAPIProfileRouters(a *Web, e *echo.Echo) {
	logrus.WithField("route", "oapi").Debug("routes initialization")

	midlewares := []oapi.StrictMiddlewareFunc{
		ValidateStructMiddeware,
		AuthMiddeware(a.app, []string{
			"DeleteProfile",
			"GetProfile",
			"PostProfileValidate",
			"PostProfileValidateSend",
			"PatchProfilePreferences",
			"PatchProfilePassword",
			"PatchProfilePhoto",
			"DeleteProfilePhoto",
			"PatchProfileColor",
			"DeleteProfileNotifications",
			"PostProfileNotificationsTaskUUIDStar",
			"DeleteProfileNotificationsTaskUUIDStar",
			"GetProfileNotifications",
			"GetProfileInvite",
			"PostProfileInviteUUID",
			"PatchProfileFio",
			"PatchProfilePhone",
			"PostProfileLike",
			"PostProfileDislike",
			"GetProfileLikes",
			"GetProfileLogout",
			"PostProfileNotificationsTaskUUIDHide",
		}),
	}

	handlers := oapi.NewStrictHandler(a, midlewares)
	oapi.RegisterHandlers(e, handlers)
}

func (a *Web) PostProfile(_ context.Context, request oapi.PostProfileRequestObject) (oapi.PostProfileResponseObject, error) {
	user := domain.NewUser(
		request.Body.Name,
		request.Body.Lname,
		request.Body.Pname,
		request.Body.Email,
		request.Body.Phone,
		request.Body.Password,
	)

	uid, code, err := a.app.ProfileService.CreateUser(user, true)
	if err != nil {
		return nil, err
	}

	logrus.Debugf("[uuid:%s][code:%s] PostProfileRegister: user created", uid, code)

	message, err := emails.NewConfirmationMessage(code)
	if err != nil {
		return nil, err
	}

	err = a.app.EmailService.SendEmail([]string{request.Body.Email}, message)
	if err != nil {
		return nil, err
	}

	hint := "check email for code"
	if a.app.IsDev() {
		hint = code
	}

	return oapi.PostProfile200JSONResponse{
		Body: oapi.UUIDResponse{
			Uuid: uid,
		},
		Headers: oapi.PostProfile200ResponseHeaders{
			Hint: hint,
		},
	}, nil
}

func (a *Web) PostProfileLogin(_ context.Context, request oapi.PostProfileLoginRequestObject) (response oapi.PostProfileLoginResponseObject, err error) {
	uud, access, refresh, exp, err := a.app.ProfileService.Login(request.Body.Email, request.Body.Password, request.Body.RememberMe, a.app.JWT)
	if err != nil {
		return response, err
	}

	tokenCookie := a.app.JWT.GenerateTokenCookie(access, refresh, exp)

	return oapi.PostProfileLogin200JSONResponse{
		Body: oapi.ProfileLoginResponse{
			Uuid:         uud,
			AccessToken:  access,
			RefreshToken: refresh,
			ValidUntil:   exp,
		},
		Headers: oapi.PostProfileLogin200ResponseHeaders{
			SetCookie: tokenCookie.String(),
		},
	}, nil
}

func (a *Web) PostProfileLoginAs(_ context.Context, request oapi.PostProfileLoginAsRequestObject) (response oapi.PostProfileLoginAsResponseObject, err error) {
	// @todo: add admin validation
	uud, access, refresh, exp, err := a.app.ProfileService.LoginAs(request.Body.Email, false, a.app.JWT)
	if err != nil {
		return response, err
	}

	tokenCookie := a.app.JWT.GenerateTokenCookie(access, refresh, exp)

	return oapi.PostProfileLoginAs200JSONResponse{
		Body: oapi.ProfileLoginResponse{
			Uuid:         uud,
			AccessToken:  access,
			RefreshToken: refresh,
			ValidUntil:   exp,
		},
		Headers: oapi.PostProfileLoginAs200ResponseHeaders{
			SetCookie: tokenCookie.String(),
		},
	}, nil
}

func (a *Web) PostProfileValidate(_ context.Context, request oapi.PostProfileValidateRequestObject) (oapi.PostProfileValidateResponseObject, error) {
	err := a.app.ProfileService.ValidateUser(request.Body.Code)
	if err != nil {
		return nil, err
	}

	return oapi.PostProfileValidate200Response{}, nil
}

func (a *Web) PostProfileValidateSimple(ctx context.Context, request oapi.PostProfileValidateSimpleRequestObject) (oapi.PostProfileValidateSimpleResponseObject, error) {
	err := a.app.ProfileService.ValidateSimple(ctx, request.Body.Email, fmt.Sprintf("%d", request.Body.Code))
	if err != nil {
		return nil, err
	}

	return oapi.PostProfileValidateSimple200Response{}, nil
}

func (a *Web) PostProfileValidateSend(_ context.Context, request oapi.PostProfileValidateSendRequestObject) (oapi.PostProfileValidateSendResponseObject, error) {
	code, err := a.app.ProfileService.SentValidate(request.Body.Email)
	if err != nil {
		return nil, err
	}

	message, err := emails.NewConfirmationMessage(code)
	if err != nil {
		return nil, err
	}

	err = a.app.EmailService.SendEmail([]string{request.Body.Email}, message)
	if err != nil {
		return nil, err
	}

	hint := "check email for code"
	if a.app.IsDev() {
		hint = code
	}

	return oapi.PostProfileValidateSend200Response{
		Headers: oapi.PostProfileValidateSend200ResponseHeaders{
			Hint: hint,
		},
	}, nil
}

func (a *Web) PostProfileValidateSimpleSend(ctx context.Context, request oapi.PostProfileValidateSimpleSendRequestObject) (oapi.PostProfileValidateSimpleSendResponseObject, error) {
	code, err := a.app.ProfileService.SentValidateSimple(ctx, request.Body.Email)
	if err != nil {
		return nil, err
	}

	message, err := emails.NewConfirmationMessage(code)
	if err != nil {
		return nil, err
	}

	err = a.app.EmailService.SendEmail([]string{request.Body.Email}, message)
	if err != nil {
		return nil, err
	}

	hint := "check email for code"
	if a.app.IsDev() {
		hint = code
	}

	return oapi.PostProfileValidateSimpleSend200Response{
		Headers: oapi.PostProfileValidateSimpleSend200ResponseHeaders{
			Hint: hint,
		},
	}, nil
}

func (a *Web) GetProfile(ctx context.Context, _ oapi.GetProfileRequestObject) (oapi.GetProfileResponseObject, error) {
	claims, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	defer Span(NewSpan(ctx, "GetProfile"))()

	federations, err := a.app.FederationService.GetFederationsByUser(ctx, claims.UUID)
	if err != nil {
		return nil, err
	}

	groups, err := a.app.FederationService.GetUserGroups(ctx, claims.UUID)
	if err != nil {
		return nil, err
	}

	companies, err := a.app.FederationService.GetCompaniesByUser(ctx, claims.UUID)
	if err != nil {
		return nil, err
	}

	projects, err := a.app.FederationService.GetProjectsByUser(ctx, claims.UUID)
	if err != nil {
		return nil, err
	}

	dm, err := a.app.ProfileService.GetUser(ctx, claims.UUID, "uuid", "is_valid", "name", "lname", "pname", "phone", "email", "color", "has_photo", "preferences", "created_at", "updated_at")
	if err != nil {
		return nil, err
	}

	// Notifications
	notificationsTotal, err := a.app.NotificationsService.Count(ctx, claims.Email)
	if err != nil {
		return nil, err
	}

	// Likes
	projectLikes, err := a.app.ProfileService.GetUserLike(claims.UUID, string(profile.Project))
	if err != nil {
		return nil, err
	}
	federationLikes, err := a.app.ProfileService.GetUserLike(claims.UUID, string(profile.Federation))
	if err != nil {
		return nil, err
	}
	companyLikes, err := a.app.ProfileService.GetUserLike(claims.UUID, string(profile.Company))
	if err != nil {
		return nil, err
	}

	federationsDTO := lo.Map(federations, func(item domain.Federation, i int) dto.FederationDTOs {
		isLiked := helpers.InArray(item.UUID, federationLikes)
		return dto.FederationDTOs{
			UUID:    item.UUID,
			Name:    item.Name,
			IsLiked: &isLiked,
		}
	})

	companiesDTO := lo.Map(companies, func(item domain.Company, i int) dto.UserCompanyDTOs {
		isLiked := helpers.InArray(item.UUID, companyLikes)
		return dto.UserCompanyDTOs{
			Position: "менеджер",
			Company: dto.CompanyDTOs{
				UUID:           item.UUID,
				Name:           item.Name,
				IsLiked:        &isLiked,
				FederationUUID: &item.FederationUUID,
			},
		}
	})

	projectsDTO := lo.Map(projects, func(item domain.Project, i int) dto.ProjectDTOs {
		isLiked := helpers.InArray(item.UUID, projectLikes)
		return dto.ProjectDTOs{
			UUID:           item.UUID,
			Name:           item.Name,
			Description:    item.Description,
			IsLiked:        &isLiked,
			StatusCode:     item.StatusCode,
			FederationUUID: item.FederationUUID,
			CompanyUUID:    item.CompanyUUID,
		}
	})

	surveys, err := a.app.ProfileService.GetSurveyByUserUUID(claims.UUID)
	if err != nil {
		logrus.Error(err)
	}
	surveysDTOs := lo.Map(surveys, func(item domain.Survey, i int) dto.SurveyDTOs {
		return dto.SurveyDTOs{
			UUID:      item.UUID,
			Name:      item.Name,
			CreatedAt: item.CreatedAt,
		}
	})

	pr := *dto.NewProfileDto(
		dm,
		federationsDTO,
		companiesDTO,
		projectsDTO,
		notificationsTotal,
		groups,
		surveysDTOs,
		dm.Preferences,
	)

	return oapi.GetProfile200JSONResponse(pr), nil
}

func (a *Web) DeleteProfile(ctx context.Context, _ oapi.DeleteProfileRequestObject) (oapi.DeleteProfileResponseObject, error) {
	claims, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	err := a.app.ProfileService.DeleteUser(claims.UUID)
	if err != nil {
		return nil, err
	}

	tokenCookie := a.app.JWT.GenerateTokenCookie("", "", time.Now())

	return oapi.DeleteProfile200Response{
		Headers: oapi.DeleteProfile200ResponseHeaders{
			SetCookie: tokenCookie.String(),
		},
	}, nil
}

func (a *Web) PatchProfilePassword(ctx context.Context, request oapi.PatchProfilePasswordRequestObject) (oapi.PatchProfilePasswordResponseObject, error) {
	claims, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	err := a.app.ProfileService.ChangePassword(claims.UUID, request.Body.NewPassword, request.Body.Password)
	if err != nil {
		return nil, err
	}

	return oapi.PatchProfilePassword200Response{}, nil
}

func (a *Web) PatchProfilePhoto(ctx context.Context, request oapi.PatchProfilePhotoRequestObject) (oapi.PatchProfilePhotoResponseObject, error) {
	claims, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	defer Span(NewSpan(ctx, "PatchProfilePhoto"))()

	file, err := request.Body.NextPart()
	if errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("file is required: %w", err)
	}
	defer file.Close()

	storeFilePath := "/tmp/" + helpers.FakeString(10) + "-" + file.FileName()
	dst, err := os.Create(storeFilePath)
	if err != nil {
		logrus.Errorf("error creating file: %s", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		logrus.Error(err)
	}

	logrus.Debug("file saved to disk:", storeFilePath)

	err = a.app.ProfileService.UploadPhoto(ctx, storeFilePath, claims.UUID)
	if err != nil {
		return nil, err
	}

	defer os.Remove(storeFilePath)

	return oapi.PatchProfilePhoto200JSONResponse(dto.ProfilePhotoDTO{
		Small:  a.app.ProfileService.GetSmallPhoto(claims.UUID),
		Medium: a.app.ProfileService.GetMediumPhoto(claims.UUID),
		Large:  a.app.ProfileService.GetLargePhoto(claims.UUID),
	}), nil
}

func (a *Web) DeleteProfilePhoto(ctx context.Context, _ oapi.DeleteProfilePhotoRequestObject) (oapi.DeleteProfilePhotoResponseObject, error) {
	claims, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	err := a.app.ProfileService.DeletePhoto(claims.UUID)
	if err != nil {
		return nil, err
	}

	return oapi.DeleteProfilePhoto200Response{}, nil
}

func (a *Web) PatchProfileColor(ctx context.Context, request oapi.PatchProfileColorRequestObject) (oapi.PatchProfileColorResponseObject, error) {
	claims, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	err := a.app.ProfileService.ChangeColor(claims.UUID, request.Body.Color)
	if err != nil {
		return nil, err
	}

	return oapi.PatchProfileColor200Response{}, nil
}

func (a *Web) GetProfileInvite(ctx context.Context, _ oapi.GetProfileInviteRequestObject) (oapi.GetProfileInviteResponseObject, error) {
	claims, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	dms, err := a.app.ProfileService.GetInvites(claims.Email)
	if err != nil {
		return nil, err
	}

	dtos := lo.Map(dms, func(item domain.Invite, index int) dto.InviteDTO {
		var company *dto.CompanyDTOs
		var federation *dto.FederationDTOs
		if item.CompanyUUID != nil {
			companyDto, f := a.app.DictionaryService.FindCompany(*item.CompanyUUID)
			if f {
				company = &dto.CompanyDTOs{
					UUID: companyDto.UUID,
					Name: companyDto.Name,
				}
			}
		}

		federationDto, f := a.app.DictionaryService.FindFederation(item.FederationUUID)
		if f {
			federation = &dto.FederationDTOs{
				UUID: federationDto.UUID,
				Name: federationDto.Name,
			}
		}

		return dto.InviteDTO{
			UUID:           item.UUID,
			Email:          item.Email,
			FederationUUID: item.FederationUUID,
			CompanyUUID:    item.CompanyUUID,
			Federation:     federation,
			Company:        company,

			CreatedAt: item.CreatedAt,
		}
	})

	return oapi.GetProfileInvite200JSONResponse{
		Count: len(dms),
		Items: dtos,
	}, nil
}

func (a *Web) PatchProfileInviteUUIDAccept(_ context.Context, request oapi.PatchProfileInviteUUIDAcceptRequestObject) (oapi.PatchProfileInviteUUIDAcceptResponseObject, error) {
	err := a.app.ProfileService.AcceptInvite(request.UUID)
	if err != nil {
		return nil, err
	}

	return oapi.PatchProfileInviteUUIDAccept200Response{}, nil
}

func (a *Web) PatchProfileInviteUUIDDecline(_ context.Context, request oapi.PatchProfileInviteUUIDDeclineRequestObject) (oapi.PatchProfileInviteUUIDDeclineResponseObject, error) {
	err := a.app.ProfileService.DeclineInvite(request.UUID)
	if err != nil {
		return nil, err
	}

	return oapi.PatchProfileInviteUUIDDecline200Response{}, nil
}

func (a *Web) PatchProfileFio(ctx context.Context, request oapi.PatchProfileFioRequestObject) (oapi.PatchProfileFioResponseObject, error) {
	claims, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	err := a.app.ProfileService.ChangeFIO(claims.UUID, request.Body.Name, request.Body.Lname, request.Body.Pname)
	if err != nil {
		return nil, err
	}

	return oapi.PatchProfileFio200Response{}, nil
}

func (a *Web) PatchProfilePhone(ctx context.Context, request oapi.PatchProfilePhoneRequestObject) (oapi.PatchProfilePhoneResponseObject, error) {
	claims, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	err := a.app.ProfileService.ChangePhone(claims.UUID, request.Body.Phone)
	if err != nil {
		return nil, err
	}

	return oapi.PatchProfilePhone200Response{}, nil
}

// PostProfileLike implements oapi.StrictServerInterface.
func (a *Web) PostProfileLike(ctx context.Context, request oapi.PostProfileLikeRequestObject) (oapi.PostProfileLikeResponseObject, error) {
	claims, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	err := a.app.ProfileService.Like(claims.UUID, profile.LikeEntityType(request.Body.Type), request.Body.Uuid)
	if err != nil {
		return nil, err
	}

	return oapi.PostProfileLike200Response{}, nil
}

func (a *Web) PostProfileDislike(ctx context.Context, request oapi.PostProfileDislikeRequestObject) (oapi.PostProfileDislikeResponseObject, error) {
	claims, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	err := a.app.ProfileService.Dislike(claims.UUID, profile.LikeEntityType(request.Body.Type), request.Body.Uuid)
	if err != nil {
		return nil, err
	}

	return oapi.PostProfileDislike200Response{}, nil
}

func (a *Web) GetProfileLikes(ctx context.Context, _ oapi.GetProfileLikesRequestObject) (oapi.GetProfileLikesResponseObject, error) {
	claims, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	projects, err := a.app.ProfileService.GetUserLike(claims.UUID, string(profile.Project))
	if err != nil {
		return nil, err
	}

	tasks, err := a.app.ProfileService.GetUserLike(claims.UUID, string(profile.Task))
	if err != nil {
		return nil, err
	}

	federations, err := a.app.ProfileService.GetUserLike(claims.UUID, string(profile.Federation))
	if err != nil {
		return nil, err
	}

	companies, err := a.app.ProfileService.GetUserLike(claims.UUID, string(profile.Company))
	if err != nil {
		return nil, err
	}

	return oapi.GetProfileLikes200JSONResponse{
		Projects:    projects,
		Tasks:       tasks,
		Federations: federations,
		Companies:   companies,
	}, err
}

func (a *Web) GetProfileLogout(_ context.Context, _ oapi.GetProfileLogoutRequestObject) (oapi.GetProfileLogoutResponseObject, error) {
	tokenCookie := &http.Cookie{
		Name:     "TOKEN",
		Value:    "",
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		Path:     "/",
	}

	return oapi.GetProfileLogout200Response{
		Headers: oapi.GetProfileLogout200ResponseHeaders{
			SetCookie: tokenCookie.String(),
		},
	}, nil
}

func (a *Web) PostProfileResetSend(_ context.Context, request oapi.PostProfileResetSendRequestObject) (oapi.PostProfileResetSendResponseObject, error) {
	code, err := a.app.ProfileService.SentReset(request.Body.Email)
	if err != nil {
		return nil, err
	}

	message, err := emails.NewConfirmationMessage(code)
	if err != nil {
		return nil, err
	}

	err = a.app.EmailService.SendEmail([]string{request.Body.Email}, message)
	if err != nil {
		return nil, err
	}

	hint := "check email for reset code"
	if a.app.IsDev() {
		hint = code
	}

	return oapi.PostProfileResetSend200Response{
		Headers: oapi.PostProfileResetSend200ResponseHeaders{
			Hint: hint,
		},
	}, nil
}

func (a *Web) PostProfileReset(_ context.Context, request oapi.PostProfileResetRequestObject) (oapi.PostProfileResetResponseObject, error) {
	err := a.app.ProfileService.ResetUserPassword(request.Body.Code, request.Body.Password)
	if err != nil {
		return nil, err
	}

	return oapi.PostProfileReset200Response{}, nil
}

func (a *Web) PatchProfilePreferences(ctx context.Context, request oapi.PatchProfilePreferencesRequestObject) (oapi.PatchProfilePreferencesResponseObject, error) {
	claims, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	if request.Body == nil {
		return nil, errors.New("preferences is nil")
	}

	err := a.app.ProfileService.ChangePreferences(claims.UUID, domain.ProfilePreferences{
		Timezone: request.Body.Timezone,
	})
	if err != nil {
		return nil, err
	}

	return oapi.PatchProfilePreferences200Response{}, nil
}
