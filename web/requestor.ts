import axios from "axios";
// export const BASE_URL = 'http://alb-0y0w4z2gd7tcsjgg07.cn-hongkong.alb.aliyuncs.com';// guohao
export const BASE_URL =
	"http://alb-ktxkm8u7il6kab7qrs.cn-hongkong.alb.aliyuncs.com"; // beer

const token = "daniel@tenant1";

const instance = axios.create({
	baseURL: BASE_URL,
	headers: {
		Authorization: `Bearer ${token}`,
	},
});

export default instance;
