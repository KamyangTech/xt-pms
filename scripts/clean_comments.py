#!/usr/bin/env python3
"""清理低价值注释（复述型 / 翻译型）。

用法:
  python3 scripts/clean_comments.py            # dry-run：仅统计，不改动文件
  python3 scripts/clean_comments.py --apply    # 实际删除并打印被修改的文件

删除规则:
  A. godoc 翻译型: 紧贴 func/type/var/const 的 '// Foo foo的中文直译'
  B. 行内复述型:   以 获取/查询/转换/遍历/初始化/... 等动词开头的独立注释行

保留:
  - 含 why 关键词的注释（避免/由于/用于/注意/泄漏/缓存/兼容/告警...）
  - 含说明、举例但承载设计信息的注释（映射表、关联路径、业务规则等）
  - 所有代码本身（本脚本只删注释行）
"""
import re, os, sys

ROOTS = ["server", "web/src"]
APPLY = "--apply" in sys.argv

VERBS = ["获取","查询","读取","写入","创建","新建","生成","转换","解析","校验","验证",
         "判断","检查","遍历","循环","返回","设置","初始化","加载","保存","删除","添加",
         "更新","计算","组装","构建","处理","过滤","排序","映射","拼接","提取","清空",
         "重置","赋值","调用","发送","接收","绑定","注册","启动","停止","打印","输出",
         "收集","合并","拆分"]
VERB_RE = re.compile(r"^\s*//\s*(" + "|".join(VERBS) + ")")

WHY_RE = re.compile(r"(避免|因为|由于|用于|注意|原因|如果|重要|必须|建议|默认|示例|说明|否则|当|前提|确保|防止|兼容|修复|遗留|历史|环境变量|缓存|内存|告警|中间件|签名|替代|泄漏|边界|限制|约束|强制|依赖|耗时|性能|并发|线程|安全|兜底|降级|回滚|异常)")

def collect_files():
    out = []
    for root in ROOTS:
        for dp, _, fns in os.walk(root):
            for fn in fns:
                if fn.endswith((".go", ".ts", ".vue")):
                    out.append(os.path.join(dp, fn))
    return out

def process(path, lines, removed_a, removed_b, touched):
    out = []
    i = 0
    n = len(lines)
    changed = False
    while i < n:
        line = lines[i]
        if re.match(r"^\s*//", line):
            block = []
            j = i
            while j < n and re.match(r"^\s*//", lines[j]):
                block.append(lines[j])
                j += 1
            nxt = lines[j] if j < n else ""
            decl = re.match(r"^\s*(func|type|var|const)\s+([A-Za-z_]\w*)", nxt)
            new_block = []
            for c in block:
                ctext = c.strip().lstrip("/").strip()
                if decl and len(block) == 1:
                    name = decl.group(2)
                    if re.match(r"^" + re.escape(name) + r"\s+[\u4e00-\u9fff]+$", ctext) and not WHY_RE.search(ctext):
                        removed_a.append((path, c.rstrip("\n")))
                        changed = True
                        continue
                if WHY_RE.search(ctext):
                    new_block.append(c)
                    continue
                if VERB_RE.match(c) and not WHY_RE.search(ctext):
                    removed_b.append((path, c.rstrip("\n")))
                    changed = True
                    continue
                new_block.append(c)
            if new_block:
                out.extend(new_block)
            i = j
            continue
        out.append(line)
        i += 1
    if changed:
        touched.append(path)
    return out

files = collect_files()
removed_a, removed_b, touched = [], [], []
for path in files:
    with open(path, encoding="utf-8") as f:
        lines = f.readlines()
    res = process(path, lines, removed_a, removed_b, touched)
    if APPLY and len(res) != len(lines):
        with open(path, "w", encoding="utf-8") as f:
            f.writelines(res)

print(f"扫描文件数: {len(files)}")
print(f"A. godoc 翻译型 待删除: {len(removed_a)}")
print(f"B. 行内复述型   待删除: {len(removed_b)}")
if APPLY:
    print(f"实际修改文件数: {len(touched)}")
    print("CLEANED:")
    for t in touched:
        print(t)
