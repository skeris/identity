package identity_svc

import (
	"context"
	"github.com/themakers/identity/identity"
	"github.com/themakers/identity/identity_svc/identity_proto"
	"github.com/themakers/session"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

//go:generate protoc -I ../identity-proto ../identity-proto/identity.proto --go_out=plugins=grpc:./identity_proto

const (
	UserIDName = "user_id"
)

const SessionTokenName = "session_token"

type IdentitySvc struct {
	mgr *identity.Manager
}

func New(backend identity.Backend, sessMgr *session.Manager, identities []identity.Identity, verifiers []identity.Verifier) (*IdentitySvc, error) {
	is := &IdentitySvc{}

	if mgr, err := identity.New(
		backend,
		sessMgr,
		identities,
		verifiers,
	); err != nil {
		return nil, err
	} else {
		is.mgr = mgr
	}

	return is, nil
}

func (is *IdentitySvc) Register(public, private *grpc.Server) {
	identity_proto.RegisterIdentityServer(public, &PublicIdentityService{
		is: is,
	})
}

////////////////////////////////////////////////////////////////
//// Helpers
////

func statusError(err error) error {
	return status.Errorf(codes.Internal, "%s", err.Error())
}

////////////////////////////////////////////////////////////////
//// PublicIdentityService
////

type PublicIdentityService struct {
	is *IdentitySvc
}

func (pis *PublicIdentityService) InitializeStaticVerifier(ctx context.Context, req *identity_proto.InitializeStaticVerifierReq) (resp *identity_proto.InitializeStaticVerifierResp, err error) {
	return

}

func (pis *PublicIdentityService) Logout(ctx context.Context, req *identity_proto.LogoutReq) (resp *identity_proto.Status, err error) {
	return
}

func (pis *PublicIdentityService) UserMerge(ctx context.Context, req *identity_proto.UserMergeReq) (resp *identity_proto.UserMergeResp, err error) {
	return
}

func (pis *PublicIdentityService) StartVerification(ctx context.Context, req *identity_proto.StartVerificationReq) (resp *identity_proto.StartVerificationResp, err error) {
	//resp := &identity_proto.StartVerificationResp{}
	sess := pis.is.mgr.Session(ctx)
	defer sess.Dispose()
	vd := []identity.VerifierData{}
	code, idnn := pis.is.mgr.StartVerification(req.Identity, req.VerifierName, ctx, vd)

	//resp := &identity_proto.StartVerificationResp{}

	return &identity_proto.StartVerificationResp{IdentityName: idnn, VerifierName: req.VerifierName, VerificationCode: code}, nil
}

func (pis *PublicIdentityService) CancelAuthentication(ctx context.Context, req *identity_proto.CancelAuthenticationReq) (resp *identity_proto.Status, err error) {
	return
}

func (pis *PublicIdentityService) StartAuthentication(ctx context.Context, req *identity_proto.StartAuthenticationReq) (resp *identity_proto.StartAuthenticationResp, err error) {
	sess := pis.is.mgr.Session(ctx)
	defer sess.Dispose()

	sessToken := pis.is.mgr.GetSessionToken(ctx)
	/*if sessToken == "" {
		panic("No session")
	}*/
	authres := pis.is.mgr.StartAuthentication(sessToken)
	if authres {

		verdir := make(map[string]string)
		return &identity_proto.StartAuthenticationResp{VerificationDirections: verdir}, nil
	}
	return &identity_proto.StartAuthenticationResp{}, nil

}

func (pis *PublicIdentityService) ListMyIdentitiesAndVerifiers(ctx context.Context, u *identity_proto.MyVerifiersDetailRequest) (response *identity_proto.VerifierDetailsResponse, err error) {
	resp := &identity_proto.VerifierDetailsResponse{}
	idns, vers := pis.is.mgr.ListMyIdentitiesAndVerifiers(u.Identity)
	for _, ver := range vers {
		resp.Verifiers = append(resp.Verifiers, &identity_proto.VerifierDetails{
			Name:           ver.Name,
			SupportRegular: ver.SupportRegular,
			SupportReverse: ver.SupportReverse,
			SupportOAuth2:  ver.SupportOAuth2,
			SupportStatic:  ver.SupportStatic,
		})
	}
	for _, idn := range idns {
		resp.IdentitiyNames = append(resp.IdentitiyNames, idn.Name)
	}

	return

}

func (pis *PublicIdentityService) ListIdentitiesAndVerifiers(ctx context.Context, q *identity_proto.VerifiersDetailsRequest) (response *identity_proto.VerifierDetailsResponse, err error) {
	sess := pis.is.mgr.Session(ctx)
	defer sess.Dispose()

	resp := &identity_proto.VerifierDetailsResponse{}
	idns, vers := pis.is.mgr.ListAllIndentitiesAndVerifiers()

	for _, ver := range vers {
		resp.Verifiers = append(resp.Verifiers, &identity_proto.VerifierDetails{
			Name:           ver.Name,
			SupportRegular: ver.SupportRegular,
			SupportReverse: ver.SupportReverse,
			SupportOAuth2:  ver.SupportOAuth2,
			SupportStatic:  ver.SupportStatic,
		})
	}
	for _, idn := range idns {
		resp.IdentitiyNames = append(resp.IdentitiyNames, idn.Name)
	}

	return resp, nil
}

func (pis *PublicIdentityService) Verify(ctx context.Context, req *identity_proto.VerifyReq) (resp *identity_proto.VerifyResp, err error) {
	//TODO get session and user
	sess := pis.is.mgr.Session(ctx)
	defer sess.Dispose()
	resp = &identity_proto.VerifyResp{}
	code := pis.is.mgr.GetVerificationCode(pis.is.mgr.GetSessionToken(ctx), req.VerifierName)
	if code == req.VerificationCode {
		resp.VerifyStatus = true
	} else {
		resp.VerifyStatus = false
	}

	return resp, nil
}

func (pis *PublicIdentityService) CheckStatus(ctx context.Context, r *identity_proto.StatusReq) (*identity_proto.Status, error) {
	// todo finish get status
	sess := pis.is.mgr.Session(ctx)
	defer sess.Dispose()
	resp := &identity_proto.Status{}

	sessionToken := pis.is.mgr.GetSessionToken(ctx)
	/*if sessionToken == "" {
		panic("No session")
	}*/
	authentication, err := pis.is.mgr.GetStatus(sessionToken)
	if err != nil {
		panic(err)
	}
	if authentication.FactorsCount != 0 {
		resp.Authenticated = true
	} else {
		resp.Authenticated = false
	}

	return resp, nil

}

////////////////////////////////////////////////////////////////
//// PrivateAuthenticationService
////

type PrivateAuthenticationService struct {
	auth *IdentitySvc
}

////////////////////////////////////////////////////
///// Helpers

func GetSessionTokenFromContext(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		// todo modificate to empty context
		panic(ok)
	}

	if at := md.Get(SessionTokenName); len(at) != 0 {
		return at[0]
	} else {
		return ""
	}

}
