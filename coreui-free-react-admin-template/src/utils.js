export function downloadCSV(data) {
  if (!data || data.length === 0) return;
  // 取出所有字段名
  const header = Object.keys(data[0]);
  // 生成 CSV 行：第一行为表头
  const csvRows = [];
  csvRows.push(header.join(','));
  
  data.forEach((row) => {
    const values = header.map((fieldName) => {
      let val = row[fieldName];
      // 如果是 cookies_json，直接输出原始字符串，不再额外包裹双引号
      if (fieldName === "cookies_json") {
        // 替换内部的换行符或逗号，避免 CSV 格式错误（可根据需要调整）
        // 例如：将换行符替换为空格
        return `"${val.replace(/[\r\n]+/g, ' ')}"`;
      } else {
        // 对其它字段，使用 JSON.stringify 以防内容中含有逗号或特殊字符
        return JSON.stringify(val !== null ? val : '');
      }
    });
    csvRows.push(values.join(','));
  });
  
  const csvContent = csvRows.join('\r\n');
  const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
  const url = URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.setAttribute('href', url);
  link.setAttribute('download', `sessions_${Date.now()}.csv`);
  link.style.visibility = 'hidden';
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
}
