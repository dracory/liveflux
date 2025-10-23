package main

import (
    "sort"
    "strings"
    "sync"
)

type User struct {
    ID    int
    Name  string
    Email string
    Role  string
}

type userRepo struct {
    mu    sync.RWMutex
    users map[int]*User
    next  int
}

var repo = newUserRepo()

func newUserRepo() *userRepo {
    r := &userRepo{users: map[int]*User{}}
    r.createUnsafe("Alice Johnson", "alice@acme.io", "Design")
    r.createUnsafe("Brian Cooper", "brian@acme.io", "Engineering")
    r.createUnsafe("Clara Mills", "clara@acme.io", "Product Management")
    r.createUnsafe("Diego Santos", "diego@acme.io", "Customer Success")
    return r
}

func (r *userRepo) createUnsafe(name, email, role string) *User {
    r.next++
    u := &User{ID: r.next, Name: name, Email: email, Role: role}
    r.users[u.ID] = u
    return u
}

func (r *userRepo) List() []User {
    r.mu.RLock()
    defer r.mu.RUnlock()
    out := make([]User, 0, len(r.users))
    for _, u := range r.users {
        out = append(out, *u)
    }
    sort.Slice(out, func(i, j int) bool {
        if strings.EqualFold(out[i].Name, out[j].Name) {
            return out[i].ID < out[j].ID
        }
        return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name)
    })
    return out
}

func (r *userRepo) Create(name, email, role string) *User {
    r.mu.Lock()
    defer r.mu.Unlock()
    name = sanitizeName(name)
    email = strings.TrimSpace(email)
    role = canonicalRole(role)
    r.next++
    u := &User{ID: r.next, Name: name, Email: email, Role: role}
    r.users[u.ID] = u
    return u
}

func (r *userRepo) Update(id int, name, email, role string) (*User, bool) {
    r.mu.Lock()
    defer r.mu.Unlock()
    u, ok := r.users[id]
    if !ok {
        return nil, false
    }
    u.Name = sanitizeName(name)
    u.Email = strings.TrimSpace(email)
    u.Role = canonicalRole(role)
    return &User{ID: u.ID, Name: u.Name, Email: u.Email, Role: u.Role}, true
}

func (r *userRepo) Delete(id int) (*User, bool) {
    r.mu.Lock()
    defer r.mu.Unlock()
    u, ok := r.users[id]
    if !ok {
        return nil, false
    }
    delete(r.users, id)
    return &User{ID: u.ID, Name: u.Name, Email: u.Email, Role: u.Role}, true
}

func sanitizeName(name string) string {
    name = strings.TrimSpace(name)
    if name == "" {
        return "Unnamed"
    }
    return name
}

func canonicalRole(role string) string {
    role = strings.TrimSpace(role)
    if role == "" {
        return "Engineering"
    }
    parts := strings.Fields(role)
    for i, p := range parts {
        lower := strings.ToLower(p)
        if len(lower) == 0 {
            continue
        }
        parts[i] = strings.ToUpper(lower[:1]) + lower[1:]
    }
    return strings.Join(parts, " ")
}
