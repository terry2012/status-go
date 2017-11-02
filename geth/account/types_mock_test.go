// Code generated by MockGen. DO NOT EDIT.
// Source: geth/account/accounts.go

// Package account is a generated GoMock package.
package account

import (
	ecdsa "crypto/ecdsa"
	accounts "github.com/ethereum/go-ethereum/accounts"
	keystore "github.com/ethereum/go-ethereum/accounts/keystore"
	common "github.com/ethereum/go-ethereum/common"
	gomock "github.com/golang/mock/gomock"
	extkeys "github.com/status-im/status-go/extkeys"
	reflect "reflect"
)

// MockaccountNode is a mock of accountNode interface
type MockaccountNode struct {
	ctrl     *gomock.Controller
	recorder *MockaccountNodeMockRecorder
}

// MockaccountNodeMockRecorder is the mock recorder for MockaccountNode
type MockaccountNodeMockRecorder struct {
	mock *MockaccountNode
}

// NewMockaccountNode creates a new mock instance
func NewMockaccountNode(ctrl *gomock.Controller) *MockaccountNode {
	mock := &MockaccountNode{ctrl: ctrl}
	mock.recorder = &MockaccountNodeMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockaccountNode) EXPECT() *MockaccountNodeMockRecorder {
	return m.recorder
}

// AccountKeyStore mocks base method
func (m *MockaccountNode) AccountKeyStore() (accountKeyStorer, error) {
	ret := m.ctrl.Call(m, "AccountKeyStore")
	ret0, _ := ret[0].(accountKeyStorer)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AccountKeyStore indicates an expected call of AccountKeyStore
func (mr *MockaccountNodeMockRecorder) AccountKeyStore() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AccountKeyStore", reflect.TypeOf((*MockaccountNode)(nil).AccountKeyStore))
}

// AccountManager mocks base method
func (m *MockaccountNode) AccountManager() (gethAccountManager, error) {
	ret := m.ctrl.Call(m, "AccountManager")
	ret0, _ := ret[0].(gethAccountManager)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AccountManager indicates an expected call of AccountManager
func (mr *MockaccountNodeMockRecorder) AccountManager() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AccountManager", reflect.TypeOf((*MockaccountNode)(nil).AccountManager))
}

// WhisperService mocks base method
func (m *MockaccountNode) WhisperService() (whisperService, error) {
	ret := m.ctrl.Call(m, "WhisperService")
	ret0, _ := ret[0].(whisperService)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// WhisperService indicates an expected call of WhisperService
func (mr *MockaccountNodeMockRecorder) WhisperService() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WhisperService", reflect.TypeOf((*MockaccountNode)(nil).WhisperService))
}

// MockgethAccountManager is a mock of gethAccountManager interface
type MockgethAccountManager struct {
	ctrl     *gomock.Controller
	recorder *MockgethAccountManagerMockRecorder
}

// MockgethAccountManagerMockRecorder is the mock recorder for MockgethAccountManager
type MockgethAccountManagerMockRecorder struct {
	mock *MockgethAccountManager
}

// NewMockgethAccountManager creates a new mock instance
func NewMockgethAccountManager(ctrl *gomock.Controller) *MockgethAccountManager {
	mock := &MockgethAccountManager{ctrl: ctrl}
	mock.recorder = &MockgethAccountManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockgethAccountManager) EXPECT() *MockgethAccountManagerMockRecorder {
	return m.recorder
}

// Wallets mocks base method
func (m *MockgethAccountManager) Wallets() []accounts.Wallet {
	ret := m.ctrl.Call(m, "Wallets")
	ret0, _ := ret[0].([]accounts.Wallet)
	return ret0
}

// Wallets indicates an expected call of Wallets
func (mr *MockgethAccountManagerMockRecorder) Wallets() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Wallets", reflect.TypeOf((*MockgethAccountManager)(nil).Wallets))
}

// MockaccountKeyStorer is a mock of accountKeyStorer interface
type MockaccountKeyStorer struct {
	ctrl     *gomock.Controller
	recorder *MockaccountKeyStorerMockRecorder
}

// MockaccountKeyStorerMockRecorder is the mock recorder for MockaccountKeyStorer
type MockaccountKeyStorerMockRecorder struct {
	mock *MockaccountKeyStorer
}

// NewMockaccountKeyStorer creates a new mock instance
func NewMockaccountKeyStorer(ctrl *gomock.Controller) *MockaccountKeyStorer {
	mock := &MockaccountKeyStorer{ctrl: ctrl}
	mock.recorder = &MockaccountKeyStorerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockaccountKeyStorer) EXPECT() *MockaccountKeyStorerMockRecorder {
	return m.recorder
}

// AccountDecryptedKey mocks base method
func (m *MockaccountKeyStorer) AccountDecryptedKey(account accounts.Account, password string) (accounts.Account, *keystore.Key, error) {
	ret := m.ctrl.Call(m, "AccountDecryptedKey", account, password)
	ret0, _ := ret[0].(accounts.Account)
	ret1, _ := ret[1].(*keystore.Key)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// AccountDecryptedKey indicates an expected call of AccountDecryptedKey
func (mr *MockaccountKeyStorerMockRecorder) AccountDecryptedKey(account, password interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AccountDecryptedKey", reflect.TypeOf((*MockaccountKeyStorer)(nil).AccountDecryptedKey), account, password)
}

// IncSubAccountIndex mocks base method
func (m *MockaccountKeyStorer) IncSubAccountIndex(account accounts.Account, password string) error {
	ret := m.ctrl.Call(m, "IncSubAccountIndex", account, password)
	ret0, _ := ret[0].(error)
	return ret0
}

// IncSubAccountIndex indicates an expected call of IncSubAccountIndex
func (mr *MockaccountKeyStorerMockRecorder) IncSubAccountIndex(account, password interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IncSubAccountIndex", reflect.TypeOf((*MockaccountKeyStorer)(nil).IncSubAccountIndex), account, password)
}

// ImportExtendedKey mocks base method
func (m *MockaccountKeyStorer) ImportExtendedKey(extKey *extkeys.ExtendedKey, password string) (accounts.Account, error) {
	ret := m.ctrl.Call(m, "ImportExtendedKey", extKey, password)
	ret0, _ := ret[0].(accounts.Account)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ImportExtendedKey indicates an expected call of ImportExtendedKey
func (mr *MockaccountKeyStorerMockRecorder) ImportExtendedKey(extKey, password interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ImportExtendedKey", reflect.TypeOf((*MockaccountKeyStorer)(nil).ImportExtendedKey), extKey, password)
}

// Accounts mocks base method
func (m *MockaccountKeyStorer) Accounts() []accounts.Account {
	ret := m.ctrl.Call(m, "Accounts")
	ret0, _ := ret[0].([]accounts.Account)
	return ret0
}

// Accounts indicates an expected call of Accounts
func (mr *MockaccountKeyStorerMockRecorder) Accounts() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Accounts", reflect.TypeOf((*MockaccountKeyStorer)(nil).Accounts))
}

// MockwhisperService is a mock of whisperService interface
type MockwhisperService struct {
	ctrl     *gomock.Controller
	recorder *MockwhisperServiceMockRecorder
}

// MockwhisperServiceMockRecorder is the mock recorder for MockwhisperService
type MockwhisperServiceMockRecorder struct {
	mock *MockwhisperService
}

// NewMockwhisperService creates a new mock instance
func NewMockwhisperService(ctrl *gomock.Controller) *MockwhisperService {
	mock := &MockwhisperService{ctrl: ctrl}
	mock.recorder = &MockwhisperServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockwhisperService) EXPECT() *MockwhisperServiceMockRecorder {
	return m.recorder
}

// DeleteKeyPairs mocks base method
func (m *MockwhisperService) DeleteKeyPairs() error {
	ret := m.ctrl.Call(m, "DeleteKeyPairs")
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteKeyPairs indicates an expected call of DeleteKeyPairs
func (mr *MockwhisperServiceMockRecorder) DeleteKeyPairs() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteKeyPairs", reflect.TypeOf((*MockwhisperService)(nil).DeleteKeyPairs))
}

// SelectKeyPair mocks base method
func (m *MockwhisperService) SelectKeyPair(key *ecdsa.PrivateKey) error {
	ret := m.ctrl.Call(m, "SelectKeyPair", key)
	ret0, _ := ret[0].(error)
	return ret0
}

// SelectKeyPair indicates an expected call of SelectKeyPair
func (mr *MockwhisperServiceMockRecorder) SelectKeyPair(key interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SelectKeyPair", reflect.TypeOf((*MockwhisperService)(nil).SelectKeyPair), key)
}

// MocksubAccountFinder is a mock of subAccountFinder interface
type MocksubAccountFinder struct {
	ctrl     *gomock.Controller
	recorder *MocksubAccountFinderMockRecorder
}

// MocksubAccountFinderMockRecorder is the mock recorder for MocksubAccountFinder
type MocksubAccountFinderMockRecorder struct {
	mock *MocksubAccountFinder
}

// NewMocksubAccountFinder creates a new mock instance
func NewMocksubAccountFinder(ctrl *gomock.Controller) *MocksubAccountFinder {
	mock := &MocksubAccountFinder{ctrl: ctrl}
	mock.recorder = &MocksubAccountFinderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MocksubAccountFinder) EXPECT() *MocksubAccountFinderMockRecorder {
	return m.recorder
}

// Find mocks base method
func (m *MocksubAccountFinder) Find(keyStore accountKeyStorer, extKey *extkeys.ExtendedKey, subAccountIndex uint32) ([]accounts.Account, error) {
	ret := m.ctrl.Call(m, "Find", keyStore, extKey, subAccountIndex)
	ret0, _ := ret[0].([]accounts.Account)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Find indicates an expected call of Find
func (mr *MocksubAccountFinderMockRecorder) Find(keyStore, extKey, subAccountIndex interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Find", reflect.TypeOf((*MocksubAccountFinder)(nil).Find), keyStore, extKey, subAccountIndex)
}

// MockkeyFileFinder is a mock of keyFileFinder interface
type MockkeyFileFinder struct {
	ctrl     *gomock.Controller
	recorder *MockkeyFileFinderMockRecorder
}

// MockkeyFileFinderMockRecorder is the mock recorder for MockkeyFileFinder
type MockkeyFileFinderMockRecorder struct {
	mock *MockkeyFileFinder
}

// NewMockkeyFileFinder creates a new mock instance
func NewMockkeyFileFinder(ctrl *gomock.Controller) *MockkeyFileFinder {
	mock := &MockkeyFileFinder{ctrl: ctrl}
	mock.recorder = &MockkeyFileFinderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockkeyFileFinder) EXPECT() *MockkeyFileFinderMockRecorder {
	return m.recorder
}

// Find mocks base method
func (m *MockkeyFileFinder) Find(keyStoreDir string, addressObj common.Address) ([]byte, error) {
	ret := m.ctrl.Call(m, "Find", keyStoreDir, addressObj)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Find indicates an expected call of Find
func (mr *MockkeyFileFinderMockRecorder) Find(keyStoreDir, addressObj interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Find", reflect.TypeOf((*MockkeyFileFinder)(nil).Find), keyStoreDir, addressObj)
}

// MockextendedKeyImporter is a mock of extendedKeyImporter interface
type MockextendedKeyImporter struct {
	ctrl     *gomock.Controller
	recorder *MockextendedKeyImporterMockRecorder
}

// MockextendedKeyImporterMockRecorder is the mock recorder for MockextendedKeyImporter
type MockextendedKeyImporterMockRecorder struct {
	mock *MockextendedKeyImporter
}

// NewMockextendedKeyImporter creates a new mock instance
func NewMockextendedKeyImporter(ctrl *gomock.Controller) *MockextendedKeyImporter {
	mock := &MockextendedKeyImporter{ctrl: ctrl}
	mock.recorder = &MockextendedKeyImporterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockextendedKeyImporter) EXPECT() *MockextendedKeyImporterMockRecorder {
	return m.recorder
}

// Import mocks base method
func (m *MockextendedKeyImporter) Import(keyStore accountKeyStorer, extKey *extkeys.ExtendedKey, password string) (string, string, error) {
	ret := m.ctrl.Call(m, "Import", keyStore, extKey, password)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(string)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// Import indicates an expected call of Import
func (mr *MockextendedKeyImporterMockRecorder) Import(keyStore, extKey, password interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Import", reflect.TypeOf((*MockextendedKeyImporter)(nil).Import), keyStore, extKey, password)
}
