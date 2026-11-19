#version 300 es
precision highp float;
in vec2 v_texcoord;
uniform sampler2D tex;
out vec4 fragColor;
const mat3 seen = mat3(1.255528, -0.078411, 0.004733, -0.076749, 0.930809, 0.691367, -0.178779, 0.147602, 0.303900);
const mat3 shift = mat3(1.0, 0.0, 0.0, 0.0, 1.0, 0.0, 0.7, 0.7, 0.0);
void main() {
    vec4 pixel = texture(tex, v_texcoord);
    vec3 lost = pixel.rgb - seen * pixel.rgb;
    fragColor = vec4(clamp(pixel.rgb + shift * lost, 0.0, 1.0), pixel.a);
}
