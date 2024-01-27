import axios, { AxiosResponse } from 'axios';

/**
 *
 * @export
 * @interface LoginStatusResponse
 */
export interface LoginStatusResponse {
  /**
   *
   * @type {boolean}
   * @memberof LoginStatusResponse
   */
  isLoggedIn: boolean;

  /**
   * Error message if there's any issue.
   *
   * @type {string}
   * @memberof LoginStatusResponse
   */
  error?: string;
}

/**
 *
 * @export
 * @interface LoginResponse
 */
export interface LoginResponse {
  /**
   *
   * @type {boolean}
   * @memberof LoginResponse
   */
  isLoggedIn: boolean;
}

/**
 * @export
 * @interface RunningProcessDetail
 */
export interface RunningProcessDetail {
  /**
   * CPU使用率
   * @type {number}
   * @memberof RunningProcessDetail
   */
  cpu_percent: number;

  /**
   * 内存信息
   * @type {MemoryInfo}
   * @memberof RunningProcessDetail
   */
  memory: MemoryInfo;

  /**
   * 磁盘信息
   * @type {DiskInfo}
   * @memberof RunningProcessDetail
   */
  disk: DiskInfo;

  /**
   * 系统启动时间
   * @type {number}
   * @memberof RunningProcessDetail
   */
  boot_time: number;

  /**
   * 进程信息
   * @type {ProcessInfo}
   * @memberof RunningProcessDetail
   */
  process: ProcessInfo;
}

/**
 * @export
 * @interface MemoryInfo
 */
export interface MemoryInfo {
  /**
   * 总内存
   * @type {number}
   * @memberof MemoryInfo
   */
  total: number;

  /**
   * 可用内存
   * @type {number}
   * @memberof MemoryInfo
   */
  available: number;

  /**
   * 内存使用率
   * @type {number}
   * @memberof MemoryInfo
   */
  percent: number;
}

/**
 * @export
 * @interface DiskInfo
 */
export interface DiskInfo {
  /**
   * 磁盘总容量
   * @type {number}
   * @memberof DiskInfo
   */
  total: number;

  /**
   * 磁盘剩余空间
   * @type {number}
   * @memberof DiskInfo
   */
  free: number;

  /**
   * 磁盘使用率
   * @type {number}
   * @memberof DiskInfo
   */
  percent: number;
}

/**
 * @export
 * @interface ProcessInfo
 */
export interface ProcessInfo {
  /**
   * 当前进程ID
   * @type {number}
   * @memberof ProcessInfo
   */
  pid: number;

  /**
   * 进程状态
   * @type {string}
   * @memberof ProcessInfo
   */
  status: string;

  /**
   * 进程使用的内存
   * @type {number}
   * @memberof ProcessInfo
   */
  memory_used: number;

  /**
   * 进程CPU使用率
   * @type {number}
   * @memberof ProcessInfo
   */
  cpu_percent: number;

  /**
   * 进程启动时间
   * @type {number}
   * @memberof ProcessInfo
   */
  start_time: number;
}

class Api {
  private axiosInstance;

  constructor() {
    this.axiosInstance = axios.create({
      // Axios 实例可以不设置 baseURL，会自动使用当前网页的地址
      withCredentials: true, // 确保 cookies 随请求发送
    });
  }

  public async checkLoginStatus(): Promise<LoginStatusResponse> {
    try {
      const response: AxiosResponse<LoginStatusResponse> =
        await this.axiosInstance.get('/api/check-login-status');
      return response.data;
    } catch (error) {
      console.error('Error checking login status:', error);
      throw error;
    }
  }

  public async loginApi(
    username: string,
    password: string
  ): Promise<LoginResponse> {
    try {
      const response: AxiosResponse<LoginResponse> =
        await this.axiosInstance.post('/api/login', {
          username,
          password,
        });
      return response.data;
    } catch (error) {
      console.error('Error during login:', error);
      throw error;
    }
  }
}

const api = new Api();
export default api;
