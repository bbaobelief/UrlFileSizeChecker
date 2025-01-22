<script setup>
import { ref, onMounted } from 'vue';
import { ElMessage, ElInput, ElButton, ElProgress, ElSlider } from 'element-plus';
import { CheckFileSizeConcurrent, CancelCheck } from '../wailsjs/go/main/App';

// 使用 ref 定义响应式变量
const urlInput = ref(''); // 输入框中的 URL
const urlList = ref([]); // 存储所有 URL 及其文件大小
const progress = ref(0); // 进度条进度
const isChecking = ref(false); // 是否正在检查
const concurrency = ref(50); // 并发数，默认 50
const outputFileName = ref('output.xlsx'); // 输出文件名，默认 output.xlsx

// 检查文件大小的方法
const checkFileSize = async () => {
  if (!urlInput.value) {
    ElMessage.error('请输入URL或选择文件');
    return;
  }

  const urls = urlInput.value
    .split(/[\n,]/) // 支持换行符或逗号分隔
    .map((url) => url.trim()) // 去除空白字符
    .filter((url) => url); // 过滤空字符串

  if (urls.length === 0) {
    ElMessage.error('请输入有效的URL');
    return;
  }

  urlList.value = [];
  progress.value = 0;
  isChecking.value = true;

  try {
    const results = await CheckFileSizeConcurrent(urls, concurrency.value, outputFileName.value);
    urlList.value = results;
    ElMessage.success(`检查完成，结果已保存到 ${outputFileName.value}`);
  } catch (error) {
    if (error.message === "context canceled") {
      ElMessage.warning('检查已取消');
    } else {
      ElMessage.error('检查失败');
    }
  } finally {
    isChecking.value = false;
  }
};

// 监听进度事件
onMounted(() => {
  window.runtime.EventsOn('progress', (newProgress) => {
    progress.value = newProgress;
  });
});

// 拖拽文件上传
const handleDrop = (event) => {
  event.preventDefault();
  const file = event.dataTransfer.files[0];
  if (!file) {
    return;
  }

  // 检查文件格式
  if (!file.name.endsWith('.txt')) {
    ElMessage.error('请上传 .txt 文件');
    return;
  }

  const reader = new FileReader();
  reader.onload = (e) => {
    const content = e.target.result;
    urlInput.value = content; // 将文件内容设置为输入框的值
  };
  reader.readAsText(file);
};

// 确保输出文件名以 .xlsx 结尾
const validateOutputFileName = () => {
  if (!outputFileName.value.endsWith('.xlsx')) {
    outputFileName.value = outputFileName.value.split('.')[0] + '.xlsx';
  }
};

// 取消检查
const cancelCheck = async () => {
  try {
    await CancelCheck();
    ElMessage.warning('检查已取消');
  } catch (error) {
    ElMessage.error('取消检查失败');
  }
};
</script>

<template>
  <div class="container">
    <!-- 输入框 -->
    <el-input
      v-model="urlInput"
      placeholder="请输入URL，多个URL可以用换行或逗号分隔，或将 .txt 文件拖拽到此"
      type="textarea"
      :rows="5"
      class="input-url"
      @drop="handleDrop"
      @dragover.prevent
    ></el-input>

    <!-- 设置区域 -->
    <div class="settings-container">
      <!-- 并发数和输出文件名放在同一行 -->
      <div class="settings-row">
        <div class="setting-item">
          <span>并发数：</span>
          <el-slider v-model="concurrency" :min="1" :max="100" class="slider" :disabled="isChecking"></el-slider>
          <span class="concurrency-value">{{ concurrency }}</span>
        </div>
        <div class="setting-item right-aligned">
          <span>输出文件名：</span>
          <el-input
            v-model="outputFileName"
            placeholder="请输入输出文件名"
            class="output-file-input"
            @blur="validateOutputFileName"
            :disabled="isChecking"
          ></el-input>
        </div>
      </div>
    </div>

    <!-- 按钮和进度条 -->
    <div class="button-container">
      <el-button type="primary" @click="checkFileSize" :disabled="isChecking" class="check-button">
        {{ isChecking ? '检查中...' : '检查文件' }}
      </el-button>
      <el-button type="danger" @click="cancelCheck" :disabled="!isChecking" class="cancel-button">
        取消检查
      </el-button>
    </div>

    <!-- 进度条 -->
    <el-progress
      v-if="isChecking"
      :percentage="progress"
      status="success"
      class="progress-bar"
    ></el-progress>
  </div>
</template>

<style scoped>
.container {
  padding: 20px;
  max-width: 1000px;
  margin: 0 auto;
}

.input-url {
  box-sizing: border-box;
  margin-bottom: 20px;
  padding: 12px;
  font-size: 14px;
  border-radius: 6px;
  border: 1px solid #ddd;
  background-color: #fff;
  transition: all 0.3s;
}

.input-url:focus {
  border-color: #409eff;
  outline: none;
}

.settings-container {
  margin-bottom: 20px;
}

.settings-row {
  display: flex;
  justify-content: space-between;
  gap: 20px;
  flex-wrap: wrap;
}

.setting-item {
  display: flex;
  align-items: center;
  gap: 10px;
  flex: 1;
}

.right-aligned {
  justify-content: flex-end;
}

.slider {
  width: 150px;
}

.concurrency-value {
  min-width: 30px;
  text-align: center;
  font-weight: bold;
}

.output-file-input {
  max-width: 250px;
}

.button-container {
  display: flex;
  gap: 20px;
  margin-bottom: 20px;
}

.check-button,
.cancel-button {
  flex: 1;
  height: 40px;
}

.check-button {
  background-color: #409eff;
  border-color: #409eff;
  color: #fff;
  transition: background-color 0.3s;
}

.check-button:hover {
  background-color: #66b1ff;
}

.cancel-button {
  background-color: #f56c6c;
  border-color: #f56c6c;
  color: #fff;
  transition: background-color 0.3s;
}

.cancel-button:hover {
  background-color: #f79b9b;
}

.progress-bar {
  margin-top: 20px;
}

@media (max-width: 768px) {
  .settings-row {
    flex-direction: column;
  }

  .output-file-input,
  .slider {
    width: 100%;
  }
}
</style>
