package service

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
)

func TestBuildDeptTree(t *testing.T) {
	depts := []*entity.SysDept{
		{ID: 2, DeptName: "生产部", ParentID: 1, Sort: 0, Status: 1, CreateTime: time.Now(), UpdateTime: time.Now()},
		{ID: 3, DeptName: "技术部", ParentID: 11, Sort: 0, Status: 1, CreateTime: time.Now(), UpdateTime: time.Now()},
		{ID: 11, DeptName: "总经办", ParentID: 0, Sort: 0, Status: 1, CreateTime: time.Now(), UpdateTime: time.Now()},
	}
	tree := BuildDeptTree(depts)
	b, _ := json.MarshalIndent(tree, "", "  ")
	fmt.Println(string(b))
	
	if len(tree) != 2 {
		t.Fatalf("期望2个根节点，得到 %d", len(tree))
	}
	for _, root := range tree {
		if root.ID == 11 {
			if len(root.Children) != 1 {
				t.Fatalf("总经办期望1个子节点，得到 %d", len(root.Children))
			}
			if root.Children[0].DeptName != "技术部" {
				t.Fatalf("子节点期望为'技术部'，得到 '%s'", root.Children[0].DeptName)
			}
		}
	}
	fmt.Println("=== TREE BUILD OK ===")
}
