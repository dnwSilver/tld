using System.Runtime.Serialization;
using YamlDotNet.Serialization;

namespace TUI.Providers.Dependencies;

[DataContract]
[YamlSerializable]
public class HubDto
{
    [YamlMember] [DataMember] public string Origin { get; set; }

    [YamlMember] [DataMember] public string Type { get; set; }

    [YamlMember] [DataMember] public IEnumerable<ProjectDto> Projects { get; set; }
}