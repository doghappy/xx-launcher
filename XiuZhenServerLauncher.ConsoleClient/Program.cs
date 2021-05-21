using Newtonsoft.Json;
using System;
using System.Net.Sockets;
using System.Text;

namespace XiuZhenServerLauncher.ConsoleClient
{
    class Program
    {
        static void Main(string[] args)
        {
            var socket = new Socket(SocketType.Stream, ProtocolType.Tcp)
            {
                ReceiveTimeout = 3000
            };
            socket.Connect("127.0.0.1", 9599);
            int id = 4;
            int data = 1;
            if (args.Length > 0)
                int.TryParse(args[0], out id);
            if (args.Length > 1)
                int.TryParse(args[1], out data);
            string json = JsonConvert.SerializeObject(new
            {
                Id = id,
                Data = data
            });
            byte[] bytes = Encoding.UTF8.GetBytes(json);
            socket.Send(bytes);
            Console.WriteLine(json);
            try
            {
                byte[] buffer = new byte[128];
                int count = socket.Receive(buffer);
                Console.WriteLine(Encoding.UTF8.GetString(buffer, 0, count));
            }
            catch (Exception e)
            {
                Console.WriteLine(e.ToString());
            }
            socket.Disconnect(false);
        }
    }
}
