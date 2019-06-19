package mock

import (
	"context"
	"github.com/lt90s/rfschub-server/account/store"
	"strconv"
	"sync"
	"time"
)

type mockStore struct {
	mu       sync.RWMutex
	id       int
	accounts map[string]accountInfo
}

type accountInfo struct {
	id        int
	name      string
	email     string
	password  string
	createdAt int64
}

func NewMockStore() *mockStore {
	return &mockStore{
		id:       100000,
		accounts: make(map[string]accountInfo),
	}
}

func (m *mockStore) CreateAccount(ctx context.Context, name, email, password string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.accounts[name]; ok {
		return store.ErrNameUsed
	}

	for _, info := range m.accounts {
		if info.email == email {
			return store.ErrEmailRegistered
		}
	}

	m.accounts[name] = accountInfo{
		id:        m.id,
		name:      name,
		email:     email,
		password:  password,
		createdAt: time.Now().Unix(),
	}
	m.id += 1
	return nil
}

func (m *mockStore) LoginAccount(ctx context.Context, name, email, password string) (info store.AccountInfo, err error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	account, ok := m.accounts[name]
	if !ok {
		for _, a := range m.accounts {
			if a.email == email {
				account = a
				ok = true
				break
			}
		}
	}

	if !ok || account.password != password {
		err = store.ErrNoMatch
		return
	}

	info.Id = strconv.Itoa(account.id)
	info.Name = account.name
	info.CreatedAt = account.createdAt
	return
}

func (m *mockStore) GetAccountId(ctx context.Context, name string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	account, ok := m.accounts[name]
	if !ok {
		return "", store.ErrNoAccount
	}
	return strconv.Itoa(account.id), nil
}

func (m *mockStore) GetAccountsBasicInfo(ctx context.Context, uids []string) (infos []store.BasicInfo, err error) {
	return
}
