package backend_test

import (
	"context"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/skeris/identity/identity"
	"testing"
	"time"
)

func Test(t *testing.T, instantiate func(context.Context) (identity.Backend, func(context.Context) error, error)) {
	Convey("Start testing by cleaning up a database", t, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		back, cleanup, err := instantiate(ctx)
		So(back, ShouldNotBeNil)
		So(err, ShouldBeNil)

		err = cleanup(ctx)
		So(err, ShouldBeNil)

		defer (func() {
			if err := cleanup(ctx); err != nil {
				panic(err)
			}
		})()

		Convey("Should not find non-existing user", func() {
			user, err := back.GetUser(ctx, "nouser")
			So(err, ShouldBeNil)
			So(user, ShouldBeNil)
		})

		Convey("Then create a user", func() {
			userID := "uid1"
			user := &identity.User{
				ID: userID,
			}

			user, err := back.CreateUser(ctx, user)
			So(err, ShouldBeNil)
			So(user.Version, ShouldEqual, 1)

			Convey("User now should exists", func() {
				user, err := back.GetUser(ctx, userID)
				So(err, ShouldBeNil)
				So(user, ShouldNotBeNil)
				So(user.ID, ShouldEqual, userID)
				So(user.Version, ShouldEqual, 1)

				Convey("Then update user", func() {
					idn1 := identity.IdentityData{
						Name:     "N1",
						Identity: "I1",
					}
					idn2 := identity.IdentityData{
						Name:     "N2",
						Identity: "I2",
					}
					user.Identities = append(user.Identities, idn1, idn2)
					user, err := back.SaveUser(ctx, user)
					So(err, ShouldBeNil)
					So(user, ShouldNotBeNil)
					So(user.Version, ShouldEqual, 2)

					Convey("And try to find it by identity", func() {
						user, err := back.GetUserByIdentity(ctx, idn1.Name, idn1.Identity)
						So(err, ShouldBeNil)
						So(user, ShouldNotBeNil) // FIXME *
						So(user.Version, ShouldEqual, 2)
					})

					Convey("And try to find it by WRONG identity", func() {
						user, err := back.GetUserByIdentity(ctx, idn1.Name, idn2.Identity)
						So(err, ShouldBeNil)
						So(user, ShouldBeNil)
					})

					Convey("Then try to update user with wrong version", func() {
						user.AuthFactorsNumber = 3
						user.Version--
						user, err := back.SaveUser(ctx, user)
						So(err, ShouldNotBeNil)
						So(user, ShouldBeNil)
					})
				})
			})
		})

		Convey("Should not find non-existing authentication", func() {
			auth, err := back.GetAuthentication(ctx, "noauth")
			So(err, ShouldBeNil)
			So(auth, ShouldBeNil)
		})

		Convey("Then create an authentication", func() {
			authID := "aid1"
			userID := "uid1"
			auth, err := back.CreateAuthentication(ctx, authID, identity.ObjectiveSignIn, userID)
			So(err, ShouldBeNil)
			So(auth.ID, ShouldEqual, authID)
			So(auth.Objective, ShouldEqual, identity.ObjectiveSignIn)
			So(auth.UserID, ShouldEqual, userID)
			So(auth.Version, ShouldEqual, 1)

			Convey("Authentication now should exists", func() {
				auth, err := back.GetAuthentication(ctx, authID)
				So(err, ShouldBeNil)
				So(auth.ID, ShouldEqual, authID)
				So(auth.Objective, ShouldEqual, identity.ObjectiveSignIn)
				So(auth.UserID, ShouldEqual, userID)
				So(auth.Version, ShouldEqual, 1)

				Convey("Then update authentication", func() {
					auth.RequiredFactorsCount = 2
					auth, err := back.SaveAuthentication(ctx, auth)
					So(err, ShouldBeNil)
					So(auth.RequiredFactorsCount, ShouldEqual, 2)
					So(auth.Version, ShouldEqual, 2)

					Convey("Then try to update authentication with wrong version", func() {
						auth.RequiredFactorsCount = 3
						auth.Version--
						auth, err := back.SaveAuthentication(ctx, auth)
						So(err, ShouldNotBeNil)
						So(auth, ShouldBeNil)
					})

					Convey("Then try to delete authentication", func() {
						err := back.RemoveAuthentication(ctx, authID)
						So(err, ShouldBeNil)

						Convey("Should not find deleted authentication", func() {
							auth, err := back.GetAuthentication(ctx, authID)
							So(err, ShouldBeNil)
							So(auth, ShouldBeNil)
						})
					})
				})
			})
		})
	})
}
