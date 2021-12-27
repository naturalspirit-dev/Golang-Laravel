package session

import (
	"github.com/qbhy/goal/contracts"
	"github.com/qbhy/goal/utils"
	"strings"
)

type Session struct {
	id         string
	name       string
	started    bool
	attributes map[string]string
	request    contracts.HttpRequest
}

func New(name, id string, request contracts.HttpRequest) contracts.Session {
	return &Session{
		id:         id,
		name:       name,
		started:    false,
		request:    request,
		attributes: map[string]string{},
	}
}

func (this *Session) GetName() string {
	return this.name
}

func (this *Session) SetName(name string) {
	this.name = name
}

func (this *Session) GetId() string {
	return this.id
}

func (this *Session) SetId(id string) {
	this.id = id
}

func (this *Session) Start() bool {
	if !this.Has("_token") {
		this.RegenerateToken()
	}
	this.started = true
	return true
}

func (this *Session) loadSession() {

	for _, cookie := range this.request.Cookies() {
		if index := strings.Index(cookie.Name, this.name); index == 0 {
			this.attributes[strings.ReplaceAll(cookie.Name, this.name, "")] = cookie.Value
		}
	}

}

func (this *Session) Save() {
}

func (this *Session) All() map[string]string {
	return this.attributes
}

func (this *Session) Exists(key string) bool {
	_, exists := this.attributes[key]
	return exists
}

func (this *Session) Has(key string) bool {
	value, exists := this.attributes[key]
	return exists && value != ""
}

func (this *Session) Get(key, defaultValue string) string {
	value, exists := this.attributes[key]
	if !exists || value == "" {
		return defaultValue
	}
	return value
}

func (this *Session) Pull(key, defaultValue string) string {
	value, exists := this.attributes[key]
	if !exists || value == "" {
		return defaultValue
	}
	delete(this.attributes, key)
	return value
}

func (this *Session) Put(key, value string) {
	this.attributes[key] = value
}

func (this *Session) Token() string {
	return this.Get("_token", "")
}

func (this *Session) RegenerateToken() {
	this.Put("_token", utils.RandStr(40))
}

func (this *Session) Remove(key string) string {
	return this.Pull(key, "")
}

func (this *Session) Forget(keys ...string) {
	for _, key := range keys {
		delete(this.attributes, key)
	}
}

func (this *Session) Flush() {
	this.attributes = make(map[string]string)
}

func (this *Session) Invalidate() bool {
	this.Flush()
	return this.Migrate(true)
}

func (this *Session) Regenerate(destroy bool) bool {
	if !this.Migrate(destroy) {
		this.RegenerateToken()
	}
	return true
}

func (this *Session) Migrate(destroy bool) bool {
	if destroy {
		// todo: $this->handler->destroy($this->getId());
	}
	this.SetId(this.generateSessionId())
	return true
}

func (this *Session) IsStarted() bool {
	return this.started
}

func (this *Session) generateSessionId() string {
	return utils.RandStr(40)
}

func (this *Session) PreviousUrl() string {
	return this.Get("_previous.url", "")
}

func (this *Session) SetPreviousUrl(url string) {
	this.Put("_previous.url", url)
}