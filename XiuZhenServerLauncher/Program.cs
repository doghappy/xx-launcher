using FluentFTP;
using Newtonsoft.Json;
using SharpCompress.Archives;
using SharpCompress.Archives.Rar;
using SharpCompress.Common;
using System;
using System.Diagnostics;
using System.IO;
using System.Linq;
using System.Net;
using System.Net.Sockets;
using System.Text;
using System.Threading.Tasks;
using YamlDotNet.Serialization;

namespace XiuZhenServerLauncher
{
    class Program
    {
        static void ReadConfig()
        {
            string txt = File.ReadAllText("config.yml");
            var deserializer = new DeserializerBuilder().Build();
            config = deserializer.Deserialize<Config>(txt);
        }

        static void RecordLog()
        {
            string path = DateTime.Now.ToString("yyyy-MM-dd-HH-mm-ss") + ".log";
            var writer = new ConsoleTextWriter(Console.Out, path);
            Console.SetOut(writer);
        }

        static Config config;

        static async Task Main(string[] args)
        {
            ReadConfig();
            RecordLog();

            var server = new TcpListener(IPAddress.Any, 9599);
            server.Start();
            Console.WriteLine("【修真服务启动器】已开启：" + server.LocalEndpoint.ToString());
            while (true)
            {
                var client = await server.AcceptTcpClientAsync();
                var ipEndPoint = client.Client.RemoteEndPoint as IPEndPoint;
                string ip = ipEndPoint.Address.ToString();
                if (config.Whitelist.Contains(ip))
                {
                    Console.WriteLine(ipEndPoint + " 已连接");
                    using (var stream = client.GetStream())
                    {
                        while (!(client.Client.Poll(1, SelectMode.SelectRead) && client.Client.Available == 0))
                        {
                            try
                            {
                                byte[] bytes = new byte[client.Available];
                                await stream.ReadAsync(bytes, 0, bytes.Length);
                                string data = Encoding.UTF8.GetString(bytes);
                                if (data != string.Empty)
                                {
                                    Console.WriteLine("Receive: " + data);
                                    try
                                    {
                                        var req = JsonConvert.DeserializeObject<Request<int>>(data);
                                        await ExcuteRequestAsync(req);
                                    }
                                    catch (Exception e)
                                    {
                                        Console.WriteLine(e);
                                    }
                                }
                            }
                            catch (Exception e)
                            {
                                Console.WriteLine(e);
                            }
                        }
                    }
                }
                else
                {
                    Console.WriteLine($"非白名单 ip \"{ip}\" 试图访问服务，已拒绝。");
                    client.Close();
                }
            }
        }

        static async Task ExcuteRequestAsync(Request<int> req)
        {
            switch (req.Id)
            {
                case 0:
                    StartServer(req.Data);
                    break;
                case 1:
                    StopServer(req.Data);
                    break;
                case 2:
                    await UpdateServerAsync(req.Data);
                    break;
                case 3:
                    await UpdateConfigAsync(req.Data);
                    break;
            }
        }

        static Bat GetBat(int regionId)
        {
            return config.Bats.Single(b => b.RegionId == regionId);
        }

        static void StartServer(int regionId)
        {
            Console.WriteLine("开服");
            var bat = GetBat(regionId);
            Process.Start(new ProcessStartInfo
            {
                WorkingDirectory = bat.WorkDir,
                FileName = bat.Start
            });
        }

        static void StopServer(int regionId)
        {
            Console.WriteLine("关服");
            var bat = GetBat(regionId);
            Process.Start(new ProcessStartInfo
            {
                WorkingDirectory = bat.WorkDir,
                FileName = bat.Stop
            });
        }

        static async Task UpdateServerAsync(int regionId)
        {
            var bat = GetBat(regionId);
            await DownloadDecompressAsync(bat.WorkDir, "/Game");
            Console.WriteLine("更新服务完成");
        }

        static async Task UpdateConfigAsync(int regionId)
        {
            var bat = GetBat(regionId);
            await DownloadDecompressAsync(bat.WorkDir, "/Config");
            Console.WriteLine("更新配置完成");
        }

        static async Task OnConnected(Func<FtpClient, Task> func)
        {
            using (var ftp = new FtpClient(config.Ftp.Host))
            {
                ftp.Credentials = new NetworkCredential(config.Ftp.User, config.Ftp.Password);
                await ftp.ConnectAsync();
                await func(ftp);
            }
        }

        static async Task DownloadDecompressAsync(string workDir, string ftpPath)
        {
            await OnConnected(async ftp =>
            {
                var files = await ftp.GetListingAsync(ftpPath);
                var file = files
                    .Where(f => f.Type == FtpFileSystemObjectType.File && Path.GetExtension(f.Name).ToLower() == ".rar")
                    .OrderByDescending(f => f.Modified)
                    .FirstOrDefault();
                if (file == null)
                {
                    Console.WriteLine("更新配置失败，未找到更新包。");
                }
                else
                {
                    string[] arr = file.Name.Split('.');
                    if (arr.Length > 2)
                    {
                        string search = $"*.{arr[arr.Length - 2]}.{arr[arr.Length - 1]}";
                        foreach (string f in Directory.EnumerateFiles(workDir, search))
                        {
                            File.Delete(f);
                        }
                    }

                    Console.WriteLine("正在下载：" + file.FullName);
                    string path = Path.Combine(workDir, file.Name);
                    await ftp.DownloadFileAsync(path, file.FullName);
                    using (var archive = RarArchive.Open(path))
                    {
                        foreach (var entry in archive.Entries)
                        {
                            Console.WriteLine("正在解压：" + entry.Key);
                            entry.WriteToDirectory(workDir, new ExtractionOptions
                            {
                                ExtractFullPath = true,
                                Overwrite = true
                            });
                        }
                    }
                    Console.WriteLine("解压完成");
                }
            });
        }
    }
}
