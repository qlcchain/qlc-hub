package jwt

import (
	"testing"
	"time"
)

func TestJWTManager_Verify(t *testing.T) {
	privateKey := NewBase58Key()
	duration := time.Hour * 24 * 365 // 1year
	if m, err := NewJWTManager(privateKey, duration); err == nil {
		if token, err := m.Generate(Admin); err == nil {
			if user, err := m.Verify(token); err == nil {
				if err := user.Valid(); err != nil {
					t.Fatal(err)
				}
			} else {
				t.Fatal(err)
			}
		} else {
			t.Fatal(err)
		}
	} else {
		t.Fatal(err)
	}
}

func TestJWTManager_Refresh(t *testing.T) {
	privateKey := NewBase58Key()
	duration := time.Hour * 24 * 365 // 1year
	if m, err := NewJWTManager(privateKey, duration); err == nil {
		token, err := m.Generate(User)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(token)
		if token2, err := m.Refresh(token); err != nil {
			t.Fatal(err)
		} else {
			t.Log(token2)
			if user, err := m.Verify(token2); err == nil {
				if err := user.Valid(); err != nil {
					t.Fatal(err)
				}
			} else {
				t.Fatal(err)
			}
		}
	} else {
		t.Fatal(err)
	}
}
