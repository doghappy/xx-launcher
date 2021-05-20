using System.IO;
using System;
using System.Text;

namespace XiuZhenServerLauncher
{
    sealed class ConsoleTextWriter : TextWriter
    {
        public ConsoleTextWriter(TextWriter writer, string path)
        {
            this.writer = writer;
            this.path = path;
        }

        readonly TextWriter writer;
        readonly string path;

        public override Encoding Encoding => Encoding.UTF8;

        public override void WriteLine(string value)
        {
            value = DateTime.Now + " " + value;
            writer.WriteLine(value);
            File.AppendAllText(path, value);
            File.AppendAllText(path, Environment.NewLine);
        }
    }
}
