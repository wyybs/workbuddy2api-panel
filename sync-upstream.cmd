@echo off
chcp 65001 >nul
setlocal
cd /d "%~dp0"
echo === 1/4 拉取上游 ===
git fetch upstream || goto :err
echo === 2/4 更新 main（只允许快进） ===
git checkout main || goto :err
git merge --ff-only upstream/main || goto :err
echo === 3/4 推送 main 到你的 fork ===
git push origin main || goto :err
echo === 4/4 回到 local 并变基到 main ===
git checkout local || goto :err
git rebase main || goto :err
echo.
echo 同步完成：
git status -sb
exit /b 0
:err
echo.
echo *** 出错已停止，未强制改动任何内容 ***
git status -sb
exit /b 1
