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
            var socket = new Socket(SocketType.Stream, ProtocolType.Tcp);
            socket.Connect("127.0.0.1", 9599);
            int id = 0;
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
            socket.Disconnect(false);
        }
    }
}
