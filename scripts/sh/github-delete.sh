# --- 复制以下所有行并粘贴到终端 ---
# 替换为你想要删除的版本号
VERSION_TO_DELETE="v1.0.0"

echo "将要删除 Release 和 Tag: ${VERSION_TO_DELETE}"
echo "----------------------------------------"

# 步骤 1: 删除 GitHub 上的 Release
# (这也会删除相关的 Git Tag)
echo "正在删除远程 Release..."
gh release delete "${VERSION_TO_DELETE}" --yes

# 步骤 2: 删除远程 Git Tag (gh release delete 通常会处理，但这能确保万无一失)
echo "正在删除远程 Tag..."
git push origin --delete "${VERSION_TO_DELETE}"

# 步骤 3: 删除本地 Git Tag
echo "正在删除本地 Tag..."
git tag -d "${VERSION_TO_DELETE}"

echo "----------------------------------------"
echo "✅ 操作完成！Release 和 Tag '${VERSION_TO_DELETE}' 已在远程和本地被删除。"
# --- 复制到此结束 ---