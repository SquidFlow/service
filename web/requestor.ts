export const BASE_URL = 'http://alb-ktxkm8u7il6kab7qrs.cn-hongkong.alb.aliyuncs.com';
import axios from 'axios';


const instance = axios.create({
    baseURL: BASE_URL
  });

  export default instance;