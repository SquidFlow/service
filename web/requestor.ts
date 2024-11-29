import axios from "axios";
export const BASE_URL = "http://localhost:38080"; // dev

const token = "username@tenant1";

const instance = axios.create({
  baseURL: BASE_URL,
  headers: {
    Authorization: `Bearer ${token}`,
  },
});

export default instance;
