package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
)

// Project 数据结构
type Project struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Phishlets   string `json:"phishlets"`
	Lures       string `json:"lures"`
	Domain      string `json:"domain"`
}

// 持久化文件路径（你可以修改为绝对路径或相对路径）
const projectsFile = "projects.json"

// 从文件读取项目数据
func readProjects() ([]Project, error) {
	if _, err := os.Stat(projectsFile); os.IsNotExist(err) {
		// 文件不存在，返回空数组
		return []Project{}, nil
	}
	data, err := ioutil.ReadFile(projectsFile)
	if err != nil {
		return nil, err
	}
	var projects []Project
	if err := json.Unmarshal(data, &projects); err != nil {
		return nil, err
	}
	return projects, nil
}

// 将项目数据写入文件
func writeProjects(projects []Project) error {
	data, err := json.MarshalIndent(projects, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(projectsFile, data, 0644)
}

// 全局变量：项目列表和下一个 ID
var (
	projects    []Project
	projectNext int = 1
)

// 初始化项目数据
func initProjects() error {
	p, err := readProjects()
	if err != nil {
		return err
	}
	projects = p
	// 确定下一个 ID（假设项目 ID 按递增顺序分配）
	for _, prj := range projects {
		if prj.Id >= projectNext {
			projectNext = prj.Id + 1
		}
	}
	return nil
}

// handleProjects 处理 /api/projects 接口
func handleProjects(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		respJSON(w, projects)
	case http.MethodPost:
		var p Project
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		p.Id = projectNext
		projectNext++
		projects = append(projects, p)
		if err := writeProjects(projects); err != nil {
			http.Error(w, "Failed to write projects", http.StatusInternalServerError)
			return
		}
		respJSON(w, p)
	case http.MethodPut:
		var p Project
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		updated := false
		for i, proj := range projects {
			if proj.Id == p.Id {
				projects[i] = p
				updated = true
				break
			}
		}
		if !updated {
			http.Error(w, "Project not found", http.StatusNotFound)
			return
		}
		if err := writeProjects(projects); err != nil {
			http.Error(w, "Failed to write projects", http.StatusInternalServerError)
			return
		}
		respJSON(w, p)
	case http.MethodDelete:
		idStr := r.URL.Query().Get("id")
		if idStr == "" {
			http.Error(w, "Missing id", http.StatusBadRequest)
			return
		}
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid id", http.StatusBadRequest)
			return
		}
		deleted := false
		for i, proj := range projects {
			if proj.Id == id {
				projects = append(projects[:i], projects[i+1:]...)
				deleted = true
				break
			}
		}
		if !deleted {
			http.Error(w, "Project not found", http.StatusNotFound)
			return
		}
		if err := writeProjects(projects); err != nil {
			http.Error(w, "Failed to write projects", http.StatusInternalServerError)
			return
		}
		respJSON(w, map[string]interface{}{"success": true})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
