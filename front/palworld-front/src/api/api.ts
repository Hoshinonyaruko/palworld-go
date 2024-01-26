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
